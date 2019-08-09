package txemail

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/txemail/txtypes"
	gophermail "gopkg.in/jpoehls/gophermail.v0"
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

	var m gophermail.Message
	if err := pt.Render(struct {
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
	if want := `a text body <b>`; m.Body != want {
		t.Errorf("got text body %q, want %q", m.Body, want)
	}
	if want := `a html body <span class="&lt;b&gt;" />`; m.HTMLBody != want {
		t.Errorf("got html body %q, want %q", m.HTMLBody, want)
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_922(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
