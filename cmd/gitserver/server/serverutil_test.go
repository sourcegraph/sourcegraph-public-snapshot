package server

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestProgressWriter(t *testing.T) {
	testCases := []struct {
		name   string
		writes []string
		text   string
	}{
		{
			name:   "identity",
			writes: []string{"hello"},
			text:   "hello",
		},
		{
			name:   "single write begin newline",
			writes: []string{"\nhelloworld"},
			text:   "\nhelloworld",
		},
		{
			name:   "single write contains newline",
			writes: []string{"hello\nworld"},
			text:   "hello\nworld",
		},
		{
			name:   "single write end newline",
			writes: []string{"helloworld\n"},
			text:   "helloworld\n",
		},
		{
			name:   "first write end newline",
			writes: []string{"hello\n", "world"},
			text:   "hello\nworld",
		},
		{
			name:   "second write begin newline",
			writes: []string{"hello", "\nworld"},
			text:   "hello\nworld",
		},
		{
			name:   "single write begin return",
			writes: []string{"\rhelloworld"},
			text:   "helloworld",
		},
		{
			name:   "single write contains return",
			writes: []string{"hello\rworld"},
			text:   "world",
		},
		{
			name:   "single write end return",
			writes: []string{"helloworld\r"},
			text:   "helloworld\r",
		},
		{
			name:   "first write contains return",
			writes: []string{"hel\rlo", "world"},
			text:   "loworld",
		},
		{
			name:   "first write end return",
			writes: []string{"hello\r", "world"},
			text:   "world",
		},
		{
			name:   "second write begin return",
			writes: []string{"hello", "\rworld"},
			text:   "world",
		},
		{
			name:   "second write contains return",
			writes: []string{"hello", "wor\rld"},
			text:   "ld",
		},
		{
			name:   "second write ends return",
			writes: []string{"hello", "world\r"},
			text:   "helloworld\r",
		},
		{
			name:   "third write",
			writes: []string{"hello", "world\r", "hola"},
			text:   "hola",
		},
		{
			name:   "progress one write",
			writes: []string{"progress\n1%\r20%\r100%\n"},
			text:   "progress\n100%\n",
		},
		{
			name:   "progress multiple writes",
			writes: []string{"progress\n", "1%\r", "2%\r", "100%"},
			text:   "progress\n100%",
		},
		{
			name:   "one two three four",
			writes: []string{"one\ntwotwo\nthreethreethree\rfourfourfourfour\n"},
			text:   "one\ntwotwo\nfourfourfourfour\n",
		},
		{
			name:   "real git",
			writes: []string{"Cloning into bare repository '/Users/nick/.sourcegraph/repos/github.com/nicksnyder/go-i18n/.git'...\nremote: Counting objects: 2148, done.        \nReceiving objects:   0% (1/2148)   \rReceiving objects: 100% (2148/2148), 473.65 KiB | 366.00 KiB/s, done.\nResolving deltas:   0% (0/1263)   \rResolving deltas: 100% (1263/1263), done.\n"},
			text:   "Cloning into bare repository '/Users/nick/.sourcegraph/repos/github.com/nicksnyder/go-i18n/.git'...\nremote: Counting objects: 2148, done.        \nReceiving objects: 100% (2148/2148), 473.65 KiB | 366.00 KiB/s, done.\nResolving deltas: 100% (1263/1263), done.\n",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var w progressWriter
			for _, write := range testCase.writes {
				w.Write([]byte(write))
			}
			if actual := w.String(); testCase.text != actual {
				t.Fatalf("\ngot:\n%s\nwant:\n%s\n", actual, testCase.text)
			}
		})
	}
}

func TestUpdateFileIfDifferent(t *testing.T) {
	dir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	target := filepath.Join(dir, "sg_refhash")

	write := func(content string) {
		err := ioutil.WriteFile(target, []byte(content), 0600)
		if err != nil {
			t.Fatal(err)
		}
	}
	read := func() string {
		b, err := ioutil.ReadFile(target)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	update := func(content string) bool {
		ok, err := updateFileIfDifferent(target, []byte(content))
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
