# Project Context

## Purpose

photo-import is a fast, deduplicating importer that pulls media off camera cards and queue directories into a Capture One library.

It exists because Capture One's own import is slow at terabyte scale: its whole-library duplicate scan does not stay fast as the catalog grows.
photo-import replaces that step with a content-hash lookup, organizes files on disk, and leaves catalog integration to a fast per-folder Synchronize.

The library uses referenced images, so the tool only ever moves and names files on disk — it never writes to the Capture One catalog.

## Goals

- Import is idempotent: running it twice over the same source never duplicates a file in the library.
- Import stays fast across a terabyte-scale library, where a naive content scan would not.
- Re-importing a card that still holds already-imported files is near-instant.
- The library on disk is the source of truth; the index is a rebuildable cache, never authoritative.
- The tool is safe to interrupt and re-run: a crash or Ctrl-C leaves the library and index in a recoverable state.

## Critical behavior

Deduplication is by content, not by name or path.
A file is a duplicate when its BLAKE3 content hash is already in the index; duplicates are skipped silently.

Organization is by capture date.
Each imported file lands at `YYYY/MM/YYYY-MM-DD--HH-MM-SS-<original-name>`.
The date comes from EXIF `DateTimeOriginal`, falling back to `CreateDate`; a file with no readable capture date goes to `Unsorted/`.

Files move when the source is on the same volume as the library and copy otherwise, so same-volume imports are fast and cross-volume imports leave the card intact.

Only managed media is imported (`jpg/jpeg/heic/png/gif/tif/tiff/cr2/raf/dng/crw/mov/mp4/avi`); AppleDouble `._` sidecars and the index database itself are ignored.

The index is a rebuildable BLAKE3 content-hash cache in SQLite, refreshed with `index` and required after any change made to the library outside this tool.

Each source card is stamped with a `.photo-import.toml` marker at its volume root and tracked in a skip cache.
Files already pulled from a known card are skipped by size and modification time without re-reading their contents.
Reformatting a card wipes the marker and mints a fresh volume id, orphaning the old cache entries; `media list` and `media clear` manage stale entries.
This capability is specified in `specs/media-cache/`.

`--dry-run` reports what would happen and writes nothing — no files moved, no index or volume records changed.

## Tech Stack

- Go (single static binary, distributed via the `danhorst/tap` Homebrew tap).
- SQLite for the content-hash index.
- BLAKE3 for content hashing.
- `exiftool` for capture timestamps, driven in batch and in long-running `-stay_open` daemon mode; installed as a Homebrew dependency.

## Project Conventions

- Personal DBH repo style: terse commit messages with a `Co-Authored-By` footer, sparse comments, rules-first README, one-sentence-per-line Markdown.
- Standard Go layout: `cmd/photo-import` for the CLI surface, `internal/*` packages per concern (`index`, `exif`, `hash`, `organize`, `media`, `volume`, `config`).
- Tests live beside the code they cover; behavior changes ship with tests.
- Spec-driven changes go through OpenSpec; specs in `specs/` are the current truth, proposals in `changes/`.

## Important Constraints

- The default library is a live 1TB Capture One library at `/Volumes/Photos`; runs against it are real. Test against a sandbox library with `-L`/`--db`, never the default.
- The index database defaults to a dotfile inside the library so it travels with the drive; deriving `database` from `library` is intentional.
- The tool must never modify the Capture One catalog; it only moves and names files on disk.

## External Dependencies

- `exiftool` — capture-date extraction (hard runtime dependency).
- Capture One — the downstream catalog; integration is a manual per-folder Synchronize, not automated.
- Homebrew tap `danhorst/tap` — distribution.
