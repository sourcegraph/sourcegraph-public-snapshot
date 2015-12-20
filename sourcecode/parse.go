package sourcecode

import (
	"net/url"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// Parse takes the Contents of the given TreeEntry and tokenizes them,
// adding syntax highlighting along with graph linking information.
func Parse(ctx context.Context, entrySpec sourcegraph.TreeEntrySpec, entry *vcsclient.FileWithRange) (*sourcegraph.SourceCode, error) {
	if err := sanitizeEntry(entrySpec, entry); err != nil {
		return nil, err
	}

	sourceCode := Tokenize(entry)
	refs, err := entryRefs(ctx, entrySpec, entry)
	if err != nil {
		return nil, err
	}
	for _, r := range refs {
		var defURL *url.URL
		if graph.URIEqual(entrySpec.RepoRev.URI, r.DefKey().Repo) {
			rev := entrySpec.RepoRev.Rev
			if rev == "" {
				rev = entrySpec.RepoRev.CommitID
			}
			defURL = router.Rel.URLToDefAtRev(r.DefKey(), rev)
		} else {
			defURL = router.Rel.URLToDef(r.DefKey())
		}

		for _, line := range sourceCode.Lines {
			if r.Start >= uint32(line.StartByte) && r.Start <= uint32(line.EndByte) {
				for k, t := range line.Tokens {
					if t != nil {
						start, end := uint32(t.StartByte), uint32(t.EndByte)
						if (r.Start >= start && r.Start < end) ||
							(r.End > end && r.Start < start) ||
							(r.End > start && r.End <= end) {
							if t.URL == nil {
								t.URL = make([]string, 0, 1)
							}
							t.URL = append(t.URL, defURL.String())
							t.IsDef = r.Def
							line.Tokens[k] = t
						}
					}
				}
			}
		}
	}

	numRefs := len(refs)
	sourceCode.TooManyRefs = numRefs >= maxRefs
	sourceCode.NumRefs = int32(numRefs)

	return sourceCode, nil
}

type refsSortableByStart []*graph.Ref

func (r refsSortableByStart) Len() int           { return len(r) }
func (r refsSortableByStart) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r refsSortableByStart) Less(i, j int) bool { return r[i].Start < r[j].Start }
