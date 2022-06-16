package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type FileTaskLogger struct {
	f *os.File

	errored bool
	keep    bool
}

func newTaskLogger(slug string, keep bool, dir string) (*FileTaskLogger, error) {
	prefix := "changeset-" + slug

	f, err := os.CreateTemp(dir, prefix+".*.log")
	if err != nil {
		return nil, errors.Wrapf(err, "creating temporary file with prefix %q", prefix)
	}

	return &FileTaskLogger{
		f:    f,
		keep: keep,
	}, nil
}

func (tl *FileTaskLogger) Close() error {
	if err := tl.f.Close(); err != nil {
		return err
	}

	if tl.errored || tl.keep {
		return nil
	}

	if err := os.Remove(tl.f.Name()); err != nil {
		return errors.Wrapf(err, "cleaning up log file %s", tl.f.Name())
	}

	return nil
}

func (tl *FileTaskLogger) Log(s string) {
	fmt.Fprintf(tl.f, "%s %s\n", time.Now().Format(time.RFC3339Nano), s)
}

func (tl *FileTaskLogger) Logf(format string, a ...interface{}) {
	fmt.Fprintf(tl.f, "%s "+format+"\n", append([]interface{}{time.Now().Format(time.RFC3339Nano)}, a...)...)
}

func (tl *FileTaskLogger) MarkErrored() {
	tl.errored = true
}

func (tl *FileTaskLogger) Path() string {
	return tl.f.Name()
}

func (tl *FileTaskLogger) PrefixWriter(prefix string) io.Writer {
	return &prefixWriter{tl, prefix}
}

type prefixWriter struct {
	logger *FileTaskLogger
	prefix string
}

func (pw *prefixWriter) Write(p []byte) (int, error) {
	// Don't split on the final newline in this writer, split
	// content into separate lines anyways, so lines without \n
	// at the end wouldn't print properly regardless. This fixes
	// output being separated by constant newlines.
	// Otherwise:
	// > echo Hello world; echo Hello Sourcegraph
	//
	// Hello world
	//
	// Hello Sourcegraph
	//
	t := bytes.TrimSuffix(p, []byte("\n"))
	for _, line := range bytes.Split(t, []byte("\n")) {
		pw.logger.Logf("%s | %s", pw.prefix, string(line))
	}
	return len(p), nil
}
