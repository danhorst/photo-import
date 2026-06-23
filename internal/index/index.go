// Package index stores content hashes of library files in a SQLite database,
// keyed for fast duplicate lookup and incremental rescans.
package index

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS files (
	path      TEXT PRIMARY KEY,
	size      INTEGER NOT NULL,
	mtime     INTEGER NOT NULL,
	blake3    TEXT NOT NULL,
	hashed_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_files_blake3 ON files(blake3);
`

// Index is a handle to the on-disk content-hash database.
type Index struct {
	db *sql.DB
}

// Open opens (creating if needed) the index at path and applies the schema.
func Open(path string) (*Index, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	return &Index{db: db}, nil
}

// Close closes the underlying database.
func (i *Index) Close() error { return i.db.Close() }

// Lookup returns a stored path that has the given content hash, if any.
func (i *Index) Lookup(hash string) (path string, found bool, err error) {
	err = i.db.QueryRow(`SELECT path FROM files WHERE blake3 = ? LIMIT 1`, hash).Scan(&path)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return path, true, nil
}

// Cached returns the stored hash for path when its size and mtime are unchanged,
// letting a rescan skip re-hashing unmodified files.
func (i *Index) Cached(path string, size, mtime int64) (hash string, ok bool) {
	var s, m int64
	err := i.db.QueryRow(`SELECT size, mtime, blake3 FROM files WHERE path = ?`, path).Scan(&s, &m, &hash)
	if err != nil || s != size || m != mtime {
		return "", false
	}
	return hash, true
}

// Put inserts or updates the record for path.
func (i *Index) Put(path string, size, mtime int64, hash string) error {
	_, err := i.db.Exec(
		`INSERT INTO files(path, size, mtime, blake3, hashed_at) VALUES(?, ?, ?, ?, ?)
		 ON CONFLICT(path) DO UPDATE SET
		   size=excluded.size, mtime=excluded.mtime,
		   blake3=excluded.blake3, hashed_at=excluded.hashed_at`,
		path, size, mtime, hash, time.Now().Unix(),
	)
	return err
}

// Stats returns the number of indexed files and the most recent hash time.
func (i *Index) Stats() (count int64, last time.Time, err error) {
	var lastUnix sql.NullInt64
	err = i.db.QueryRow(`SELECT COUNT(*), MAX(hashed_at) FROM files`).Scan(&count, &lastUnix)
	if lastUnix.Valid {
		last = time.Unix(lastUnix.Int64, 0)
	}
	return count, last, err
}
