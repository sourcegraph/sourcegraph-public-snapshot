package local

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestMarkdownService_Render(t *testing.T) {
	var s markdown
	ctx, _ := testContext()

	tests := []struct {
		Markdown         string
		HTML             string
		EnableCheckboxes bool
		Checklist        sourcegraph.Checklist
	}{{
		`- [ ] simple
- [ ] list`,
		`<ul>
<li><input type="checkbox" class="comment-checkbox"> simple</li>
<li><input type="checkbox" class="comment-checkbox"> list</li>
</ul>
`,
		true,
		sourcegraph.Checklist{Todo: 2, Done: 0},
	}, {
		`* [ ] stars`,
		`<ul>
<li><input type="checkbox" class="comment-checkbox"> stars</li>
</ul>
`,
		true,
		sourcegraph.Checklist{Todo: 1, Done: 0},
	}, {
		`- [ ] - [ ] one box`,
		`<ul>
<li><input type="checkbox" class="comment-checkbox"> - [ ] one box</li>
</ul>
`,
		true,
		sourcegraph.Checklist{Todo: 1, Done: 0},
	}, {
		`- [ ] nested
 - [ ] list
  - [ ] is
 - [ ] unnesting
- [ ] now`,
		`<ul>
<li><input type="checkbox" class="comment-checkbox"> nested

<ul>
<li><input type="checkbox" class="comment-checkbox"> list</li>
<li><input type="checkbox" class="comment-checkbox"> is</li>
<li><input type="checkbox" class="comment-checkbox"> unnesting</li>
</ul></li>
<li><input type="checkbox" class="comment-checkbox"> now</li>
</ul>
`,
		true,
		sourcegraph.Checklist{Todo: 5, Done: 0},
	}, {
		`- [x] checked`,
		`<ul>
<li><input type="checkbox" class="comment-checkbox" checked> checked</li>
</ul>
`,
		true,
		sourcegraph.Checklist{Todo: 0, Done: 1},
	}, {
		`- [ ] checkboxes disabled
- [x] foo`,
		`<ul>
<li><input type="checkbox" class="comment-checkbox" disabled=true> checkboxes disabled</li>
<li><input type="checkbox" class="comment-checkbox" checked disabled=true> foo</li>
</ul>
`,
		false,
		sourcegraph.Checklist{Todo: 1, Done: 1},
	}, {`* [ ] check

* [ ] check


* [ ] check
`,
		`<ul>
<li><p><input type="checkbox" class="comment-checkbox"> check</p></li>

<li><p><input type="checkbox" class="comment-checkbox"> check</p></li>

<li><p><input type="checkbox" class="comment-checkbox"> check</p></li>
</ul>
`,
		true,
		sourcegraph.Checklist{Todo: 3, Done: 0},
	}, {
		`- <p>[ ] hi</p>`,
		`<ul>
<li><p><input type="checkbox" class="comment-checkbox"> hi</p></li>
</ul>
`,
		true,
		sourcegraph.Checklist{Todo: 1, Done: 0},
	}}

	for _, test := range tests {
		out, err := s.Render(ctx, &sourcegraph.MarkdownRenderOp{
			Markdown: []byte(test.Markdown),
			Opt:      sourcegraph.MarkdownOpt{EnableCheckboxes: test.EnableCheckboxes},
		})
		if err != nil {
			t.Fatal(err)
		}

		exp := &sourcegraph.MarkdownData{Rendered: []byte(test.HTML), Checklist: &test.Checklist}
		if actualHTML := string(out.Rendered); actualHTML != test.HTML {
			t.Errorf("got\n%s\nwanted\n%s", actualHTML, test.HTML)
		} else if !reflect.DeepEqual(out.Checklist, exp.Checklist) {
			t.Errorf("got checklist %+v but wanted %+v", out.Checklist, exp.Checklist)
		} else if !reflect.DeepEqual(out, exp) {
			t.Errorf("got %+v, wanted %+v", out, exp)
		}

	}
}
