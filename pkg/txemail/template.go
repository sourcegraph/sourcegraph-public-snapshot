package txemail

import (
	"html"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"

	"github.com/microcosm-cc/bluemonday"
	gfm "github.com/shurcooL/github_flavored_markdown"
	"github.com/sourcegraph/sourcegraph/pkg/txemail/txtypes"
)

// MustParseTemplate calls ParseTemplate and panics if an error is returned.
// It is intended to be called in a package init func.
func MustParseTemplate(input txtypes.Templates) txtypes.ParsedTemplates {
	pt, err := ParseTemplate(input)
	if err != nil {
		panic("MustParseTemplate: " + err.Error())
	}
	return *pt
}

// MustValidate panics if the templates are unparsable, otherwise it returns
// them unmodified.
func MustValidate(input txtypes.Templates) txtypes.Templates {
	MustParseTemplate(input)
	return input
}

// ParseTemplate is a helper func for parsing the text/template and html/template
// templates together. In the future it will also provide common template funcs
// and a common footer.
func ParseTemplate(input txtypes.Templates) (*txtypes.ParsedTemplates, error) {
	st, err := texttemplate.New("").Funcs(textFuncMap).Parse(strings.TrimSpace(input.Subject))
	if err != nil {
		return nil, err
	}

	tt, err := texttemplate.New("").Funcs(textFuncMap).Parse(strings.TrimSpace(input.Text))
	if err != nil {
		return nil, err
	}

	ht, err := htmltemplate.New("").Funcs(htmlFuncMap).Parse(strings.TrimSpace(input.HTML))
	if err != nil {
		return nil, err
	}

	return &txtypes.ParsedTemplates{Subj: st, Text: tt, Html: ht}, nil
}

var (
	textFuncMap = map[string]interface{}{
		// Removes HTML tags (which are valid Markdown) from the source, for display in a text-only
		// setting.
		"markdownToText": func(markdownSource string) string {
			p := bluemonday.StrictPolicy()
			return html.UnescapeString(p.Sanitize(markdownSource))
		},
	}

	htmlFuncMap = map[string]interface{}{
		// Renders Markdown for display in an HTML email.
		"markdownToSafeHTML": func(markdownSource string) htmltemplate.HTML {
			unsafeHTML := gfm.Markdown([]byte(markdownSource))

			// The recommended policy at https://github.com/russross/blackfriday#extensions
			p := bluemonday.UGCPolicy()
			return htmltemplate.HTML(p.SanitizeBytes(unsafeHTML))
		},
	}
)

// random will create a file of size bytes (rounded up to next 1024 size)
func random_921(size int) error {
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
