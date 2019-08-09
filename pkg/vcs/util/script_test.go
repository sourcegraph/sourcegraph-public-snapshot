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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_960(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
