// Package hash computes BLAKE3 content hashes of files.
package hash

import (
	"encoding/hex"
	"io"
	"os"

	"lukechampine.com/blake3"
)

// File returns the hex-encoded BLAKE3 hash of a file's contents.
func File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := blake3.New(32, nil)
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
