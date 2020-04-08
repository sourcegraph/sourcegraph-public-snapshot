package util

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestScriptFile(t *testing.T) {
	p, dir, err := ScriptFile("hi")
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Dir(p) != dir {
		t.Errorf("Expected %s to be in %s", p, dir)
	}
	// Only do remove when we get here just in case we return something
	// silly
	defer os.RemoveAll(dir)

	textPath := filepath.Join(dir, "text")
	err = WriteFileWithPermissions(textPath, []byte("Hello World!"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	var script string
	if runtime.GOOS == "windows" {
		script = "@echo off\ntype " + textPath + "\n"
	} else {
		script = "#!/bin/sh\ncat '" + textPath + "'\n"
	}
	err = WriteFileWithPermissions(p, []byte(script), 0500)
	if err != nil {
		t.Fatal(err)
	}

	out, err := exec.Command(p).CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "Hello World!" {
		t.Errorf("Unexpected output from script: %v", string(out))
	}
}
