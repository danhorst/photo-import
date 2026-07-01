package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/dbh/photo-management/internal/config"
	"github.com/dbh/photo-management/internal/export"
	"github.com/dbh/photo-management/internal/hash"
	"github.com/dbh/photo-management/internal/index"
	"github.com/dbh/photo-management/internal/media"
	"github.com/mattn/go-isatty"
	"github.com/schollz/progressbar/v3"
)

func cmdExport(args []string) error {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	lib, db, debug := commonFlags(fs)
	dryRun := fs.Bool("dry-run", false, "report derivatives without writing files or rows")
	since := fs.String("since", "", "limit to frames captured on/after this date (YYYY-MM-DD)")
	fs.Usage = func() { fmt.Print(usage) }
	if _, err := parseArgs(fs, args); err != nil {
		return err
	}

	var sinceDate time.Time
	if *since != "" {
		t, err := time.ParseInLocation("2006-01-02", *since, time.Local)
		if err != nil {
			return fmt.Errorf("--since must be YYYY-MM-DD, got %q", *since)
		}
		sinceDate = t
	}

	cfg, err := config.Load(*lib, *db)
	if err != nil {
		return err
	}
	longEdge, quality := cfg.ExportLongEdge, cfg.ExportQuality
	if longEdge == 0 {
		longEdge = export.DefaultLongEdge
	}
	if quality == 0 {
		quality = export.DefaultQuality
	}
	idx, err := index.Open(cfg.Database)
	if err != nil {
		return err
	}
	defer idx.Close()

	logf := debugLogger(*debug)
	start := time.Now()

	paths, err := collectArchive(cfg.Library)
	if err != nil {
		return err
	}
	frames := export.Group(paths)
	logf("found %d frame(s) in %d archive file(s)", len(frames), len(paths))

	type work struct {
		frame export.Frame
		src   export.Source
	}
	var jobs []work
	for _, f := range frames {
		if !sinceDate.IsZero() {
			d, ok := f.CaptureDate()
			if !ok || d.Before(sinceDate) {
				continue
			}
		}
		for _, s := range f.Sources() {
			jobs = append(jobs, work{f, s})
		}
	}

	showProgress := !*debug && !*dryRun && isatty.IsTerminal(os.Stderr.Fd())
	var bar *progressbar.ProgressBar
	if showProgress && len(jobs) > 0 {
		bar = progressbar.Default(int64(len(jobs)), "exporting")
	}

	gen := &export.Generator{LongEdge: longEdge, Quality: quality}
	defer gen.Close()

	var generated, skipped, failed int
	for _, w := range jobs {
		if bar != nil {
			bar.Add(1)
		}
		h, err := hash.File(w.src.Path)
		if err != nil {
			failed++
			logf("hash %s: %v", w.src.Path, err)
			continue
		}
		has, err := idx.HasDerivative(h)
		if err != nil {
			return err
		}
		if has {
			skipped++
			logf("skip %s (already generated)", w.src.Path)
			continue
		}
		dst := export.DestPath(cfg.Library, w.frame, w.src)
		if *dryRun {
			generated++
			logf("would export %s -> %s", w.src.Path, dst)
			continue
		}
		if err := gen.Generate(w.src, w.frame.Stem, h, dst); err != nil {
			failed++
			logf("export %s: %v", w.src.Path, err)
			continue
		}
		if err := idx.PutDerivative(h, w.frame.Stem, w.src.Kind, dst); err != nil {
			return err
		}
		generated++
		logf("exported %s -> %s", w.src.Path, dst)
	}
	if bar != nil {
		bar.Finish()
		fmt.Fprintln(os.Stderr)
	}

	verb := "Exported"
	if *dryRun {
		verb = "Would export"
	}
	fmt.Printf("%s %d HEIC(s) in %s; skipped %d already generated", verb, generated, time.Since(start).Round(time.Millisecond), skipped)
	if failed > 0 {
		fmt.Printf("; %d error(s)", failed)
	}
	fmt.Println(".")
	return nil
}

var yearDir = regexp.MustCompile(`^\d{4}$`)

// collectArchive returns the media files in the library's YYYY/MM tree,
// leaving Export/, Unsorted/, and other non-archive directories alone.
func collectArchive(library string) ([]string, error) {
	var paths []string
	years, err := os.ReadDir(library)
	if err != nil {
		return nil, err
	}
	for _, y := range years {
		if !y.IsDir() || !yearDir.MatchString(y.Name()) {
			continue
		}
		root := filepath.Join(library, y.Name())
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if media.IsExcludedDir(d.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			if media.IsMedia(d.Name()) {
				paths = append(paths, p)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return paths, nil
}
