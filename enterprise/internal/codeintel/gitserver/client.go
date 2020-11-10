package gitserver

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type Client struct {
	operations *operations
}

func New(observationContext *observation.Context) *Client {
	return &Client{
		operations: makeOperations(observationContext),
	}
}

// Head determines the tip commit of the default branch for the given repository.
func (c *Client) Head(ctx context.Context, dbStore DBStore, repositoryID int) (_ string, err error) {
	ctx, endObservation := c.operations.head.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	return execGitCommand(ctx, dbStore, repositoryID, "rev-parse", "HEAD")
}

type CommitGraphOptions struct {
	Commit string
	Limit  int
}

// CommitGraph returns the commit graph for the given repository as a mapping from a commit
// to its parents. If a commit is supplied, the returned graph will be rooted at the given
// commit. If a non-zero limit is supplied, at most that many commits will be returned.
func (c *Client) CommitGraph(ctx context.Context, dbStore DBStore, repositoryID int, opts CommitGraphOptions) (_ map[string][]string, err error) {
	ctx, endObservation := c.operations.commitGraph.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("opts.Commit", opts.Commit),
		log.Int("opts.Limit", opts.Limit),
	}})
	defer endObservation(1, observation.Args{})

	commands := []string{"log", "--all", "--pretty=%H %P"}
	if opts.Commit != "" {
		commands = append(commands, opts.Commit)
	}
	if opts.Limit > 0 {
		commands = append(commands, fmt.Sprintf("-%d", opts.Limit))
	}

	out, err := execGitCommand(ctx, dbStore, repositoryID, commands...)
	if err != nil {
		return nil, err
	}

	return parseParents(strings.Split(out, "\n")), nil
}

// parseParents converts the output of git log into a map from commits to parent commits.
// If a commit is listed but has no ancestors then its parent slice is empty but is still
// present in the map.
func parseParents(pair []string) map[string][]string {
	commits := map[string][]string{}

	for _, pair := range pair {
		line := strings.TrimSpace(pair)
		if line == "" {
			continue
		}

		parts := strings.Split(line, " ")
		commits[parts[0]] = append(commits[parts[0]], parts[1:]...)

		for _, part := range parts[1:] {
			if _, ok := commits[part]; !ok {
				commits[part] = []string{}
			}
		}
	}

	return commits
}

// RawContents returns the contents of a file in a particular commit of a repository.
func (c *Client) RawContents(ctx context.Context, dbStore DBStore, repositoryID int, commit, file string) (_ []byte, err error) {
	ctx, endObservation := c.operations.rawContents.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
		log.String("file", file),
	}})
	defer endObservation(1, observation.Args{})

	out, err := execGitCommand(ctx, dbStore, repositoryID, "show", fmt.Sprintf("%s:%s", commit, file))
	if err != nil {
		return nil, err
	}

	return []byte(out), err
}

// DirectoryChildren determines all children known to git for the given directory names via an invocation
// of git ls-tree. The keys of the resulting map are the input (unsanitized) dirnames, and the value of
// that key are the files nested under that directory.
func (c *Client) DirectoryChildren(ctx context.Context, dbStore DBStore, repositoryID int, commit string, dirnames []string) (_ map[string][]string, err error) {
	ctx, endObservation := c.operations.directoryChildren.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	out, err := execGitCommand(ctx, dbStore, repositoryID, append([]string{"ls-tree", "--name-only", commit, "--"}, cleanDirectoriesForLsTree(dirnames)...)...)
	if err != nil {
		return nil, err
	}

	return parseDirectoryChildren(dirnames, strings.Split(out, "\n")), nil
}

// cleanDirectoriesForLsTree sanitizes the input dirnames to a git ls-tree command. There are a
// few peculiarities handled here:
//
//   1. The root of the tree must be indicated with `.`, and
//   2. In order for git ls-tree to return a directory's contents, the name must end in a slash.
func cleanDirectoriesForLsTree(dirnames []string) []string {
	var args []string
	for _, dir := range dirnames {
		if dir == "" {
			args = append(args, ".")
		} else {
			if !strings.HasSuffix(dir, "/") {
				dir += "/"
			}
			args = append(args, dir)
		}
	}

	return args
}

// parseDirectoryChildren converts the flat list of files from git ls-tree into a map. The keys of the
// resulting map are the input (unsanitized) dirnames, and the value of that key are the files nested
// under that directory. If dirnames contains a directory that encloses another, then the paths will
// be placed into the key sharing the longest path prefix.
func parseDirectoryChildren(dirnames []string, paths []string) map[string][]string {
	childrenMap := map[string][]string{}

	// Ensure each directory has an entry, even if it has no children
	// listed in the gitserver output.
	for _, dirname := range dirnames {
		childrenMap[dirname] = nil
	}

	// Order directory names by length (biggest first) so that we assign
	// paths to the most specific enclosing directory in the following loop.
	sort.Slice(dirnames, func(i, j int) bool {
		return len(dirnames[i]) > len(dirnames[j])
	})

	for _, path := range paths {
		if strings.Contains(path, "/") {
			for _, dirname := range dirnames {
				if strings.HasPrefix(path, dirname) {
					childrenMap[dirname] = append(childrenMap[dirname], path)
					break
				}
			}
		} else {
			// No need to loop here. If we have a root input directory it
			// will necessarily be the last element due to the previous
			// sorting step.
			if len(dirnames) > 0 && dirnames[len(dirnames)-1] == "" {
				childrenMap[""] = append(childrenMap[""], path)
			}
		}
	}

	return childrenMap
}

// FileExists determines whether a file exists in a particular commit of a repository.
func (c *Client) FileExists(ctx context.Context, dbStore DBStore, repositoryID int, commit, file string) (_ bool, err error) {
	ctx, endObservation := c.operations.fileExists.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
		log.String("file", file),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := repositoryIDToRepo(ctx, dbStore, repositoryID)
	if err != nil {
		return false, err
	}

	if _, err := git.ResolveRevision(ctx, repo, nil, commit, git.ResolveRevisionOptions{}); err != nil {
		return false, errors.Wrap(err, "git.ResolveRevision")
	}

	if _, err := git.Stat(ctx, repo, api.CommitID(commit), file); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// ListFiles returns a list of root-relative file paths matching the given pattern in a particular
// commit of a repository.
func (c *Client) ListFiles(ctx context.Context, dbStore DBStore, repositoryID int, commit string, pattern *regexp.Regexp) (_ []string, err error) {
	ctx, endObservation := c.operations.listFiles.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
		log.String("pattern", pattern.String()),
	}})
	defer endObservation(1, observation.Args{})

	out, err := execGitCommand(ctx, dbStore, repositoryID, "ls-tree", "--name-only", "-r", commit, "--")
	if err != nil {
		return nil, err
	}

	var matching []string
	for _, path := range strings.Split(out, "\n") {
		if pattern.MatchString(path) {
			matching = append(matching, path)
		}
	}

	return matching, nil
}

// execGitCommand executes a git command for the given repository by identifier.
func execGitCommand(ctx context.Context, dbStore DBStore, repositoryID int, args ...string) (string, error) {
	repo, err := repositoryIDToRepo(ctx, dbStore, repositoryID)
	if err != nil {
		return "", err
	}

	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = repo
	out, err := cmd.CombinedOutput(ctx)
	return string(bytes.TrimSpace(out)), errors.Wrap(err, "gitserver.Command")
}

// repositoryIDToRepo creates a gitserver.Repo from a repository identifier.
func repositoryIDToRepo(ctx context.Context, dbStore DBStore, repositoryID int) (gitserver.Repo, error) {
	repoName, err := dbStore.RepoName(ctx, repositoryID)
	if err != nil {
		return gitserver.Repo{}, errors.Wrap(err, "store.RepoName")
	}

	return gitserver.Repo{Name: api.RepoName(repoName)}, nil
}
