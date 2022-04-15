package main

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/sourcegraph/codenotify"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// CodeOwnerResolver returns the list of GitHub users who subscribed themselves
// to changes on particular paths, through the OWNERS file.
//
// See https://github.com/sourcegraph/codenotify for more informations.
type CodeOwnerResolver interface {
	Resolve(ref string) ([]string, error)
}

func NewMockCodeOwnerResolver(mapping map[string][]string) CodeOwnerResolver {
	if mapping == nil {
		mapping = map[string][]string{}
	}
	return &mockCodeOwnerResolver{
		mapping: mapping,
	}
}

type mockCodeOwnerResolver struct {
	mapping map[string][]string
}

func (m *mockCodeOwnerResolver) Resolve(ref string) ([]string, error) {
	return m.mapping[ref], nil
}

func NewGitCodeOwnerResolver(cloneURL string, clonePath string) CodeOwnerResolver {
	return &gitCodeOwnerResolver{
		cloneURL:  cloneURL,
		clonePath: clonePath,
	}
}

type gitCodeOwnerResolver struct {
	cloneURL  string
	clonePath string
	once      sync.Once
}

func (g *gitCodeOwnerResolver) Resolve(ref string) ([]string, error) {
  var err error
	g.once.Do(func() {
		_, err = exec.Command("git", "clone", "--quiet", "git@github.com:sourcegraph/sourcegraph.git", g.clonePath).Output()
	})
  if err != nil {
    return nil, err 
  }

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	os.Chdir(g.clonePath)
	defer func() { _ = os.Chdir(cwd) }()

	owners := []string{}
	diff, err := exec.Command("git", "show", "--name-only", ref).CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, "error getting diff for %s", ref)
	}
	paths, err := readLines(diff)
	if err != nil {
		return nil, errors.Errorf("error scanning lines from diff %s: %s\n%s", ref, err, string(diff))
	}
	for _, path := range paths {
		fs := codenotify.NewGitFS(g.clonePath, ref)
		ownersList, err := codenotify.Subscribers(fs, path, "CODENOTIFY")
		if err != nil {
			return nil, errors.Wrapf(err, "error computing the subscribers for %s", path)
		}
		owners = append(owners, ownersList...)
	}
	return uniqAndSanitize(owners), nil
}

func uniqAndSanitize(strs []string) []string {
	set := map[string]struct{}{}
	for _, str := range strs {
		set[str] = struct{}{}
	}
	uniqStrs := make([]string, 0, len(set))
	for str := range set {
		uniqStrs = append(uniqStrs, strings.TrimLeft(str, "@"))
	}
	return uniqStrs
}

func readLines(b []byte) ([]string, error) {
	lines := []string{}
	scanner := bufio.NewScanner(bytes.NewBuffer(b))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
