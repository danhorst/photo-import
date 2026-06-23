package organize

import (
	"testing"
	"time"
)

func TestDest(t *testing.T) {
	when := time.Date(2026, 6, 2, 15, 47, 30, 0, time.Local)
	got := Dest("/Volumes/Photos", when, "DSCF1297.RAF")
	want := "/Volumes/Photos/2026/06/2026-06-02--15-47-30-DSCF1297.RAF"
	if got != want {
		t.Errorf("Dest() = %q, want %q", got, want)
	}
}

func TestDestZeroPads(t *testing.T) {
	when := time.Date(2004, 1, 9, 7, 5, 3, 0, time.Local)
	got := Dest("/lib", when, "x.jpg")
	want := "/lib/2004/01/2004-01-09--07-05-03-x.jpg"
	if got != want {
		t.Errorf("Dest() = %q, want %q", got, want)
	}
}
