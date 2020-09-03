package search

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/store"
)

type ignoreMatcher struct {
	ignoreList []string
}

var (
	lineComment = "#"
	ignoreFile  = ".sourcegraph/sourcegraphignore"
)

func ParseSourcegraphignore(r *bytes.Reader) (patterns []string, error error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		// ignore empty lines
		if len(line) == 0 {
			continue
		}
		// ignore comments
		if strings.HasPrefix(strings.TrimSpace(line), lineComment) {
			continue
		}
		// add trailing "/" to make sure we don't match files that
		// share a prefix with a directory
		if !strings.HasSuffix(line, "/") {
			line += "/"
		}

		line = strings.TrimPrefix(line, "/")
		patterns = append(patterns, line)
	}
	return patterns, scanner.Err()
}

func newIgnoreMatcher(zf *store.ZipFile) (*ignoreMatcher, error) {
	var patterns []string
	var err error
	for _, file := range zf.Files {
		if file.Name == ignoreFile {
			fileBuf := zf.DataFor(&file)
			r := bytes.NewReader(fileBuf)
			patterns, err = ParseSourcegraphignore(r)
			break
		}
	}
	return &ignoreMatcher{ignoreList: patterns}, err
}

func (ig *ignoreMatcher) match(file *store.SrcFile) bool {
	for _, pattern := range ig.ignoreList {
		if strings.HasPrefix(strings.TrimPrefix(file.Name, "/"), pattern) {
			return true
		}
	}
	return false
}
