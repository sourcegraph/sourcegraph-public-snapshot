package ctags

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/ctxvfs"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/randstring"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/ctags/parser"
)

func (h *Handler) getTags(ctx context.Context) ([]parser.Tag, error) {
	// If we are already running generateTags, we definitely don't want to run
	// it again. If we aren't running it, this mutex will be very fast.
	h.tagsMu.Lock()
	defer h.tagsMu.Unlock()
	if h.tags == nil {
		tags, err := generateTags(ctx, h.fs)
		if err != nil {
			return nil, err
		}
		h.tags = tags
	}
	return h.tags, nil
}

var ignoreFiles = []string{".srclib-cache", "node_modules", "vendor", "dist", ".git"}

// generateTags runs ctags and parses the output.
func generateTags(ctx context.Context, fs ctxvfs.FileSystem) ([]parser.Tag, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "run ctags")
	defer span.Finish()

	// Download the files to a random path, because while we trust ctags to not
	// be actively malicious, it would be pretty easy to compromise.
	randPath := randstring.NewLen(16)
	filesDir, err := ioutil.TempDir("", randPath)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(filesDir)

	err = copyRepoArchive(ctx, fs, filesDir)
	if err != nil {
		return nil, err
	}

	args := []string{"-f", "-", "--fields=*", "--excmd=pattern", "--languages=Ruby,C"}
	args = append(args, "-R", filesDir)
	excludeArgs := make([]string, 0, len(ignoreFiles))
	for _, ignoreFile := range ignoreFiles {
		excludeArgs = append(excludeArgs, fmt.Sprintf("--exclude=%s", ignoreFile))
	}
	args = append(args, excludeArgs...)
	cmd := exec.Command("ctags", args...)

	// Pipe out the ouput of ctags directly into the ctags parser
	rc, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	tags, err := parser.Parse(bufio.NewReader(rc))
	if err != nil {
		return nil, err
	}

	// Wait until ctags has finished processing the files
	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	// Strip useless path information
	for i, tag := range tags {
		tags[i].File = strings.TrimPrefix(tag.File, filesDir)
	}

	return tags, nil
}
