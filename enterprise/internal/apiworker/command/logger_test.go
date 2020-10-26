package command

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLogger(t *testing.T) {
	logger := NewLogger("baz2", "BAR2")

	for i := 0; i < 3; i++ {
		outReader, outWriter := io.Pipe()
		errReader, errWriter := io.Pipe()

		go func() {
			defer outWriter.Close()
			defer errWriter.Close()

			_, _ = io.Copy(outWriter, bytes.NewReader([]byte(fmt.Sprintf("foo%[1]d bar%[1]d\nbaz%[1]d", i+1))))
			_, _ = io.Copy(errWriter, bytes.NewReader([]byte(fmt.Sprintf("FOO%[1]d BAR%[1]d\nBAZ%[1]d", i+1))))
		}()

		logger.RecordCommand(
			[]string{"test", strconv.FormatInt(int64(i)+1, 10)},
			outReader,
			errReader,
		)
	}

	expected := `
		stderr: BAZ1
		stderr: BAZ2
		stderr: BAZ3
		stderr: FOO1 BAR1
		stderr: FOO2 ******
		stderr: FOO3 BAR3
		stdout: baz1
		stdout: ******
		stdout: baz3
		stdout: foo1 bar1
		stdout: foo2 bar2
		stdout: foo3 bar3
		test 1
		test 2
		test 3
	`
	if diff := cmp.Diff(normalizeLogs(expected), normalizeLogs(logger.String())); diff != "" {
		t.Errorf("unexpected log output (-want +got):\n%s", diff)
	}
}

func normalizeLogs(text string) (filtered []string) {
	for _, line := range strings.Split(text, "\n") {
		if line := strings.TrimSpace(line); line != "" {
			filtered = append(filtered, line)
		}
	}
	sort.Strings(filtered)

	return filtered
}
