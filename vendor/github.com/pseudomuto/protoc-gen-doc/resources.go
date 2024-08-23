package gendoc

import (
	_ "embed" // for including embedded resources
)

var (
	//go:embed resources/docbook.tmpl
	docbookTmpl []byte
	//go:embed resources/html.tmpl
	htmlTmpl []byte
	//go:embed resources/markdown.tmpl
	markdownTmpl []byte
	//go:embed resources/scalars.json
	scalarsJSON []byte
)
