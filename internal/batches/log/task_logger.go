package log

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/pkg/errors"
)

type TaskLogger struct {
	f *os.File

	errored bool
	keep    bool
}

func newTaskLogger(slug string, keep bool, dir string) (*TaskLogger, error) {
	prefix := "changeset-" + slug

	f, err := ioutil.TempFile(dir, prefix+".*.log")
	if err != nil {
		return nil, errors.Wrapf(err, "creating temporary file with prefix %q", prefix)
	}

	return &TaskLogger{
		f:    f,
		keep: keep,
	}, nil
}

func (tl *TaskLogger) Close() error {
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

func (tl *TaskLogger) Log(s string) {
	fmt.Fprintf(tl.f, "%s %s\n", time.Now().Format(time.RFC3339Nano), s)
}

func (tl *TaskLogger) Logf(format string, a ...interface{}) {
	fmt.Fprintf(tl.f, "%s "+format+"\n", append([]interface{}{time.Now().Format(time.RFC3339Nano)}, a...)...)
}

func (tl *TaskLogger) MarkErrored() {
	tl.errored = true
}

func (tl *TaskLogger) Path() string {
	return tl.f.Name()
}

func (tl *TaskLogger) PrefixWriter(prefix string) io.Writer {
	return &prefixWriter{tl, prefix}
}

type prefixWriter struct {
	logger *TaskLogger
	prefix string
}

func (pw *prefixWriter) Write(p []byte) (int, error) {
	for _, line := range bytes.Split(p, []byte("\n")) {
		pw.logger.Logf("%s | %s", pw.prefix, string(line))
	}
	return len(p), nil
}
