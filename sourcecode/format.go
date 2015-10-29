package sourcecode

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"sort"
	"text/template"

	"github.com/sourcegraph/annotate"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/conf"
)

// Format performs syntax highlighting and ref linking on entry. It
// modifies entry.Contents during these 2 transformations and returns
// other metadata about the operation.
func Format(ctx context.Context, entrySpec sourcegraph.TreeEntrySpec, entry *vcsclient.FileWithRange, highlightStrings []string) (*sourcegraph.FormatResult, error) {
	if err := sanitizeEntry(entrySpec, entry); err != nil {
		return nil, err
	}

	res := sourcegraph.FormatResult{
		LineStartByteOffsets: computeLineStartByteOffsets(entry.Contents),
	}
	startByte := uint32(entry.StartByte)

	refs, err := entryRefs(ctx, entrySpec, entry)
	if err != nil {
		return nil, err
	}
	refAnns := make([]*annotate.Annotation, len(refs))
	for i, r := range refs {
		// If iref, use the revspec of the file; otherwise, use the
		// empty revspec to assume we're pointing to the default
		// branch of the destination repo (which is not *always*
		// correct but we aren't handling that complexity yet).
		//
		// This means that iref links don't jump to a different commit
		// of the same repo, which would be weird.
		var defURL *url.URL
		if graph.URIEqual(entrySpec.RepoRev.URI, r.DefKey().Repo) {
			// TODO(sqs): when showing delta files, we want the 2nd
			// arg to be entrySpec.RepoRev.CommitID because the
			// branches change. but when showing normal files, we want
			// it to be .Rev because then the URLs are nice. there's
			// currently a bug if it's .CommitID where the JS doesn't
			// know where to jump to on the page. make it work for
			// both and make Rev/CommitID configurable in this func.
			defURL = router.Rel.URLToDefAtRev(r.DefKey(), entrySpec.RepoRev.Rev)
		} else {
			defURL = router.Rel.URLToDef(r.DefKey())
		}

		var class string // HTML <a> elem class
		if r.Def {
			class += " def"
		}
		if !defURL.IsAbs() {
			defURL = conf.AppURL(ctx).ResolveReference(defURL)
		}
		refAnns[i] = &annotate.Annotation{
			Start: int(r.Start - startByte), End: int(r.End - startByte),
			Left:  []byte(fmt.Sprintf(`<a href="%s" class="ref%s">`, template.HTMLEscapeString(defURL.String()), class)),
			Right: []byte("</a>"),
		}
	}

	var matchAnns []*annotate.Annotation
	for _, hl := range highlightStrings {
		hlb := []byte(hl)
		for i := 0; i < len(entry.Contents); {
			p := bytes.Index(entry.Contents[i:], hlb)
			if p == -1 {
				break
			}
			matchAnns = append(matchAnns, &annotate.Annotation{
				Start:     i + p,
				End:       i + p + len(hlb),
				Left:      []byte(`<span class="match">`),
				Right:     []byte("</span>"),
				WantInner: -150,
			})
			i += p + len(hlb) + 1
		}
	}

	res.NumRefs = int32(len(refAnns))
	res.TooManyRefs = res.NumRefs > maxRefs

	// Syntax highlighting.
	hlAnns, err := SyntaxHighlight(entry.Name, entry.Contents)
	if err != nil {
		return nil, err
	}
	for _, a := range hlAnns {
		a.WantInner = 100 // put syntax highlight <span> inside of ref <a> so color shows
	}

	// Combine and sort annotations.
	anns := annotate.Annotations(refAnns)
	anns = append(anns, hlAnns...)
	anns = append(anns, matchAnns...)
	sort.Sort(anns)

	// Apply annotations.
	out, err := annotate.Annotate(entry.Contents, anns, func(w io.Writer, b []byte) { template.HTMLEscape(w, b) })
	if annotate.IsOutOfBounds(err) {
		// ignore; the annotation tags are properly opened and closed by
		// Annotate anyway.
		err = nil
	}
	if err != nil {
		return nil, err
	}
	entry.Contents = out

	return &res, nil
}

func computeLineStartByteOffsets(src []byte) []int32 {
	// TODO(sqs): move this into
	// vcsstore? It already parses the file's lines using sqs/fileset.
	lines := bytes.SplitAfter(src, []byte{'\n'})
	ofs := make([]int32, len(lines))
	for i := range lines {
		if i > 0 {
			ofs[i] = ofs[i-1] + int32(len(lines[i-1]))
		}
	}
	return ofs
}
