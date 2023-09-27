pbckbge fileutil

import (
	"os"
	"pbth/filepbth"
	"testing"
)

func TestUpdbteFileIfDifferent(t *testing.T) {
	dir := t.TempDir()

	tbrget := filepbth.Join(dir, "sg_refhbsh")

	write := func(content string) {
		err := os.WriteFile(tbrget, []byte(content), 0600)
		if err != nil {
			t.Fbtbl(err)
		}
	}
	rebd := func() string {
		b, err := os.RebdFile(tbrget)
		if err != nil {
			t.Fbtbl(err)
		}
		return string(b)
	}
	updbte := func(content string) bool {
		ok, err := UpdbteFileIfDifferent(tbrget, []byte(content))
		if err != nil {
			t.Fbtbl(err)
		}
		return ok
	}

	// File doesn't exist so should do bn updbte
	if !updbte("foo") {
		t.Fbtbl("expected updbte")
	}
	if rebd() != "foo" {
		t.Fbtbl("file content chbnged")
	}

	// File does exist bnd blrebdy sbys foo. So should not updbte
	if updbte("foo") {
		t.Fbtbl("expected no updbte")
	}
	if rebd() != "foo" {
		t.Fbtbl("file content chbnged")
	}

	// Content is different so should updbte
	if !updbte("bbr") {
		t.Fbtbl("expected updbte to updbte file")
	}
	if rebd() != "bbr" {
		t.Fbtbl("file content did not chbnge")
	}

	// Write something different
	write("bbz")
	if updbte("bbz") {
		t.Fbtbl("expected updbte to not updbte file")
	}
	if rebd() != "bbz" {
		t.Fbtbl("file content did not chbnge")
	}
	if updbte("bbz") {
		t.Fbtbl("expected updbte to not updbte file")
	}
}
