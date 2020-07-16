package txemail

import (
	"testing"

	"github.com/jordan-wright/email"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
)

func TestParseTemplate(t *testing.T) {
	pt, err := ParseTemplate(txtypes.Templates{
		Subject: `{{.A}} subject {{.B}}`,
		Text: `
{{.A}} text body {{.B}}
`,
		HTML: `
{{.A}} html body <span class="{{.B}}" />
`,
	})
	if err != nil {
		t.Fatal(err)
	}

	var m email.Email
	if err := renderTemplate(pt, struct {
		A string
		B string
	}{
		A: "a",
		B: `<b>`,
	}, &m); err != nil {
		t.Fatal(err)
	}

	if want := `a subject <b>`; m.Subject != want {
		t.Errorf("got subject %q, want %q", m.Subject, want)
	}
	if want := `a text body <b>`; string(m.Text) != want {
		t.Errorf("got text body %q, want %q", string(m.Text), want)
	}
	if want := `a html body <span class="&lt;b&gt;" />`; string(m.HTML) != want {
		t.Errorf("got html body %q, want %q", string(m.HTML), want)
	}
}
