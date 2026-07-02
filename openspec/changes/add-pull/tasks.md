## 1. Reverse export

- [x] 1.1 Add an `osxphotos export` wrapper that exports assets to a queue directory with `--update`
- [x] 1.2 Apply the device model allowlist (initially `Apple iPhone 13 mini`), sourced from config so it is extensible
- [x] 1.3 Exclude any asset carrying our `catalogKey`
- [x] 1.4 Import the Live Photo still normally; ignore the motion `.mov` component
- [x] 1.5 Unit-test the allowlist / `catalogKey` filter as a plain function separate from the osxphotos shell-out

## 2. CLI

- [x] 2.1 Add `pull` to the dispatch in `cmd/pm` with `--since <date>`, `--dry-run`, and shared `--db`/`-L`
- [x] 2.2 `cmdPull` runs the export into a queue directory, then reuses the import path over it (BLAKE3 dedup + `YYYY/MM` unchanged)
- [x] 2.3 `--dry-run` reports what would be exported/imported and writes nothing

## 3. Docs

- [x] 3.1 Document `pm pull`, the device allowlist, and the two dedup layers in `README.md`
- [x] 3.2 Note the `osxphotos` dependency and Live Photos handling
- [x] 3.3 `CHANGELOG.md` entry

## 4. Verify

- [x] 4.1 `go test ./...` passes
- [x] 4.2 `openspec validate add-pull --strict` passes
- [x] 4.3 Manual: against a throwaway `--db`/`-L` sandbox and a test Photos library, pull once (imported), pull again (skipped by BLAKE3), and confirm a `catalogKey`-stamped asset is never re-ingested
