package ctags

import (
	"bufio"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/randstring"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/ctags/parser"
)

func getTags(ctx context.Context) ([]parser.Tag, error) {
	// If we are already running generateTags, we definitely don't want to run
	// it again. If we aren't running it, this mutex will be very fast.
	info := ctxInfo(ctx)
	info.tagsMu.Lock()
	defer info.tagsMu.Unlock()

	if info.tags == nil {
		tags, err := generateTags(ctx)
		if err != nil {
			return nil, err
		}
		info.tags = tags
	}
	return info.tags, nil
}

// generateTags runs ctags and parses the output.
func generateTags(ctx context.Context) ([]parser.Tag, error) {
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

	err = copyRepoArchive(ctx, filesDir)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("ctags", "-f", "-", "--fields=*", "--excmd=pattern", "-R", filesDir)

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
