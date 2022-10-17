package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateFileIfDifferent(t *testing.T) {
	dir := t.TempDir()

	target := filepath.Join(dir, "sg_refhash")

	write := func(content string) {
		err := os.WriteFile(target, []byte(content), 0600)
		if err != nil {
			t.Fatal(err)
		}
	}
	read := func() string {
		b, err := os.ReadFile(target)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	update := func(content string) bool {
		ok, err := UpdateFileIfDifferent(target, []byte(content))
		if err != nil {
			t.Fatal(err)
		}
		return ok
	}

	// File doesn't exist so should do an update
	if !update("foo") {
		t.Fatal("expected update")
	}
	if read() != "foo" {
		t.Fatal("file content changed")
	}

	// File does exist and already says foo. So should not update
	if update("foo") {
		t.Fatal("expected no update")
	}
	if read() != "foo" {
		t.Fatal("file content changed")
	}

	// Content is different so should update
	if !update("bar") {
		t.Fatal("expected update to update file")
	}
	if read() != "bar" {
		t.Fatal("file content did not change")
	}

	// Write something different
	write("baz")
	if update("baz") {
		t.Fatal("expected update to not update file")
	}
	if read() != "baz" {
		t.Fatal("file content did not change")
	}
	if update("baz") {
		t.Fatal("expected update to not update file")
	}
}
