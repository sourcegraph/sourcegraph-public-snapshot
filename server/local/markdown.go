package local

import (
	"bytes"
	"fmt"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/doc"
)

const (
	// These are the strings to look for in rendered vanilla Markdown that should be replaced with checkboxes.
	// (Vanilla markdown does not include checkboxes, so the rendering process happens in 2 steps: first render
	// vanilla markdown, then replace these strings with checkbox HTML.)
	mdUnchecked  = `<li>[ ] `
	mdUncheckedP = `<li><p>[ ] `
	mdChecked    = `<li>[x] `
	mdCheckedP   = `<li><p>[x] `
)

var Markdown sourcegraph.MarkdownServer = &markdown{}

type markdown struct{}

var _ sourcegraph.MarkdownServer = (*markdown)(nil)

func (s *markdown) Render(ctx context.Context, op *sourcegraph.MarkdownRenderOp) (*sourcegraph.MarkdownData, error) {
	rendered, err := doc.ToHTML(doc.Markdown, op.Markdown)
	if err != nil {
		return nil, err
	}

	var disabledClass = ""
	if !op.Opt.EnableCheckboxes {
		disabledClass = ` disabled=true`
	}

	htmlUnchecked := fmt.Sprintf(`<li><input type="checkbox" class="comment-checkbox"%s> `, disabledClass)
	htmlUncheckedP := fmt.Sprintf(`<li><p><input type="checkbox" class="comment-checkbox"%s> `, disabledClass)
	htmlChecked := fmt.Sprintf(`<li><input type="checkbox" class="comment-checkbox" checked%s> `, disabledClass)
	htmlCheckedP := fmt.Sprintf(`<li><p><input type="checkbox" class="comment-checkbox" checked%s> `, disabledClass)

	uncheckedCount := bytes.Count(rendered, []byte(mdUnchecked)) + bytes.Count(rendered, []byte(mdUncheckedP))
	checkedCount := bytes.Count(rendered, []byte(mdChecked)) + bytes.Count(rendered, []byte(mdCheckedP))

	rendered = bytes.Replace(rendered, []byte(mdUnchecked), []byte(htmlUnchecked), -1)
	rendered = bytes.Replace(rendered, []byte(mdUncheckedP), []byte(htmlUncheckedP), -1)
	rendered = bytes.Replace(rendered, []byte(mdChecked), []byte(htmlChecked), -1)
	rendered = bytes.Replace(rendered, []byte(mdCheckedP), []byte(htmlCheckedP), -1)

	return &sourcegraph.MarkdownData{
		Rendered: rendered,
		Checklist: &sourcegraph.Checklist{
			Todo: int32(uncheckedCount),
			Done: int32(checkedCount),
		},
	}, nil
}
