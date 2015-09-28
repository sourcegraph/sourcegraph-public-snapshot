package sourcecode

import (
	"errors"
	"path"
	"sort"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/store"
)

// ErrIsNotFile occurs when Parse or Format is called on a tree entry
// that's not a code file (e.g., a directory or symlink).
var ErrIsNotFile = errors.New("code file to format is not a file")

// Limit maxRefs to a nice round number minus one so that
// res.NumRefs is a round number if we exceed the limit. (It'd
// look weird to say "only the first 6001 refs were linked.")
const maxRefs = 5999

// entryRefs fetches all references for a given entry and spec.
func entryRefs(ctx context.Context, entrySpec sourcegraph.TreeEntrySpec, entry *vcsclient.FileWithRange) ([]*graph.Ref, error) {
	refFilters := []srcstore.RefFilter{
		srcstore.ByRepos(entrySpec.RepoRev.RepoSpec.URI),
		srcstore.ByCommitIDs(entrySpec.RepoRev.CommitID),
		srcstore.ByFiles(path.Clean(entrySpec.Path)),
		srcstore.RefFilterFunc(func(ref *graph.Ref) bool {
			return ref.Start >= uint32(entry.StartByte) && ref.End <= uint32(entry.EndByte)
		}),
	}
	refs, err := store.GraphFromContext(ctx).Refs(refFilters...)
	if err != nil {
		return nil, err
	}
	sort.Sort(refsSortableByStart(refs))
	return refs, nil
}

// sanitizeEntry checks that the passed entry and entrySpec are valid and sets unset values
// to their defaults.
func sanitizeEntry(entrySpec sourcegraph.TreeEntrySpec, entry *vcsclient.FileWithRange) error {
	if entry.Type != vcsclient.FileEntry {
		return ErrIsNotFile
	}
	if entrySpec.RepoRev.CommitID == "" {
		panic("assumes that CommitID has been resolved and stored, but it is empty (in the func that calls get, resolve the CommitID from the Rev and store it in the entrySpec's RepoRevSpec before calling get")
	}
	if entry.EndByte == 0 {
		entry.EndByte = int64(len(entry.Contents))
	}
	if entry.StartLine == 0 {
		entry.StartLine = 1
	}
	return nil
}
