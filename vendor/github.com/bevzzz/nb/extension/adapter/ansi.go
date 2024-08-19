package adapter

import (
	"io"

	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/schema"
)

// AnsiHtml wraps [ansihtml]-style function in RenderCellFunc.
//
// Usage:
//
//	extension.NewStream(
//		adapter.AnsiHtml(ansihtml.ConvertToHTML)
//	)
//
// To force ansihtml to use classes instead of inline styles, pass an anonymous function intead:
//
//	extension.NewStream(
//		adapter.AnsiHtml(func([]byte) []byte) {
//			ansihtml.ConvertToHTMLWithClasses(b, "class-", false)
//		})
//	)
//
// [ansihtml]: https://github.com/robert-nix/ansihtml
func AnsiHtml(convert func([]byte) []byte) render.RenderCellFunc {
	return func(w io.Writer, cell schema.Cell) (err error) {
		// Wrapping in <pre> helps preserve parts of the original
		// formatting such as newlines and tabs.
		io.WriteString(w, "<pre>")
		_, err = w.Write(convert(cell.Text()))
		io.WriteString(w, "</pre>")
		return
	}
}
