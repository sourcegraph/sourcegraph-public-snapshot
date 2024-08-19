package adapter

import (
	"io"

	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/schema"
)

// Blackfriday wraps [blackfriday]-style function in RenderCellFunc.
//
// Usage:
//
//	extension.NewMarkdown(
//		adapter.Blackfriday(blackfriday.MarkdownCommon)
//	)
//
// [blackfriday]: https://github.com/russross/blackfriday
func Blackfriday(convert func([]byte) []byte) render.RenderCellFunc {
	return func(w io.Writer, cell schema.Cell) (err error) {
		_, err = w.Write(convert(cell.Text()))
		return
	}
}

// Goldmark wraps [goldmark]-style function in RenderCellFunc.
//
// Usage:
//
//	extension.NewMarkdown(
//		adapter.Goldmark(func(b []byte, w io.Writer) error {
//			return goldmark.Convert(b, w, parseOptions...)
//		})
//	)
//
// Notice, how Goldmark is a bit more verbose compared to Blackfriday:
// this is because goldmark.Convert accepts variadic parser.ParseOptions, which
// is a dependency the client should capture in the closure and pass manually.
//
// [goldmark]: https://github.com/yuin/goldmark
func Goldmark(write func([]byte, io.Writer) error) render.RenderCellFunc {
	return func(w io.Writer, cell schema.Cell) error {
		return write(cell.Text(), w)
	}
}
