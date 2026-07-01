# photo-management

Fast, deduplicating photo importer for a Capture One library.

Imports media into a `YYYY/MM` tree named `YYYY-MM-DD--HH-MM-SS-<original>`, skipping files whose contents are already in the index.
Duplicate detection is a BLAKE3 content-hash lookup against a SQLite index, so it stays fast across a terabyte-scale library where Capture One's own import scan is slow.

## Install

```
brew install danhorst/tap/pm
```

Requires `exiftool`, installed automatically as a dependency.

## Use

Build the index once, and after any change made to the library outside this tool:

```
pm index
```

Import a card or queue directory:

```
pm /Volumes/UNTITLED/DCIM
```

New files are organized into the library; content duplicates are skipped silently.
The summary lists the `YYYY/MM` folders that changed — **Synchronize** those folders in Capture One to bring them into the catalog.
The library uses referenced images, so a per-folder sync is fast and sidesteps Capture One's slow whole-library duplicate scan.

Files move when the source is on the same volume as the library, and copy otherwise.
Files without a readable capture date go to `Unsorted/`.

Re-importing a card that still holds already-imported files is near-instant: each card is stamped with a `.photo-management.toml` marker at its root, and files already pulled from it are skipped by size and modification time without re-reading their contents.

## Commands

- `pm <source>` — import from a directory. Flags: `--dry-run`, `--debug`.
- `pm export` — generate presentation HEICs into `Export/` (see below). Flags: `--since YYYY-MM-DD`, `--dry-run`, `--debug`.
- `pm index` — build or refresh the content-hash index.
- `pm stats` — show index location and size.
- `pm config <cmd>` — read/write the config file (see below).
- `pm media list` — list every cached volume: name, id, file count, and last-seen date.
- `pm media clear [<id>…]` — remove a volume's skip-cache entries by id or unambiguous id prefix; with no arguments on a terminal, opens an interactive multiselect.
- `pm version` — print the version.

### Export

`pm export` turns archive frames into downsized HEIC derivatives under `Export/YYYY/MM` inside the library.
Each frame yields one base derivative — from the sibling camera JPEG, or the RAF's embedded `JpgFromRaw` when RAW-only — named `<stem>.heic`, plus one `<stem>-<suffix>.heic` per baked Capture One edit.
iPhone-origin frames (a `.HEIC` with no camera JPEG) are left alone.

Derivatives are resized to a 4096 px long edge at quality 70 (configurable via `export_long_edge` / `export_quality`), carry `DateTimeOriginal`/GPS/orientation, and are stamped with `catalogKey` (BLAKE3 of the source file) and `catalogStem` (the frame stem) in XMP.
Export is incremental: a source whose hash is already recorded in the `derivative` table is skipped, so re-runs only generate what's new.
`--since YYYY-MM-DD` scopes a run; `--dry-run` reports without writing.

`Export/` is a regenerable presentation mirror — exclude it from backup; only the master tier is backed up.

### Media cache and reformatted cards

Each card is identified by a `.photo-management.toml` marker stamped at the card's volume root.
A card stamped with the old `.photo-import.toml` marker is still recognized; new stamps use the new name.
Reformatting a card wipes that marker, so the next import treats it as a new card and mints a fresh volume id.
The old id's cache entries become orphaned: they are never read again but also never removed automatically.
Use `pm media list` to see all cached volumes and `pm media clear <id>` to remove stale entries.
Clearing touches only the index — the card and its files are unchanged — and the next import of that card re-hashes from scratch.

## Configuration

`~/.config/photo-management/photo-management.toml`:

```toml
library = "/Volumes/Photos"
database = "/Volumes/Photos/.photo-index.db"
```

Both default as shown; the database defaults to a dotfile inside the library so it travels with the drive.
An existing config at the old `~/.config/photo-import/photo-import.toml` path is still read; the first `config set` migrates it to the new path.
Override per run with `--library`/`-L` and `--db`.

Manage the file from the CLI instead of editing by hand:

```
pm config init                       # write a default config file
pm config set library /Volumes/Archive
pm config show                        # print the effective values
pm config path                        # print the file location
```

`database` derives from `library` unless set explicitly, so changing the library moves the index with it.

## Organization

Photos are renamed and organized by date.
This descends from [work by @cliss](https://gist.github.com/cliss/6854904) which, in turn, was based on a [script by Dr. Drang](http://www.leancrew.com/all-this/2013/10/photo-management-via-the-finder/).
