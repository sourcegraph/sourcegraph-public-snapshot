# nb

[![Build status](https://github.com/bevzzz/nb/actions/workflows/build.yml/badge.svg?branch=ci)](https://github.com/bevzzz/nb/actions/workflows/build.yml?query=branch%3Amain)
[![Go Reference](https://pkg.go.dev/badge/github.com/bevzzz/nb.svg)](https://pkg.go.dev/github.com/bevzzz/nb)
[![Go Report Card](https://goreportcard.com/badge/github.com/bevzzz/nb)](https://goreportcard.com/report/github.com/bevzzz/nb)
[![codecov](https://codecov.io/gh/bevzzz/nb/branch/main/graph/badge.svg?token=BS7XDXHA21)](https://codecov.io/gh/bevzzz/nb/tree/main)

	Render Jupyter Notebooks in pure Go 📔

This package is inspired by @yuin's [`goldmark`](https://github.com/yuin/goldmark) and is designed to be as clear and extensible.

The implementation follows the official [Jupyter Notebook format spec](https://nbformat.readthedocs.io/en/latest/format_description.html#the-notebook-file-format) (`nbformat`) and produces an output similar to that of [`nbconvert`](https://github.com/jupyter/nbconvert) (Jupyter's team own reference implementation) both structurally and visually. It supports all major `nbformat` schema versions: `v4.0-v4.5`, `v3.0`, `v2.0`, `v1.0`.

The package comes with an HTML renderer out of the box and can be extended to convert notebooks to other formats, such as LaTeX or PDF.

> 🏗 This package is being actively developed: its structure and APIs might change overtime.  
> If you find any bugs, please consider opening an issue or submitting a PR.

## Installation

```sh
go get github.com/bevzzz/nb
```

## Usage

`nb`'s default, no-frills converter can render markdown, code, and raw cells out of the box:

```go
b, err := os.ReadFile("notebook.ipynb")
if err != nil {
	panic(err)
}
err := nb.Convert(os.Stdout, b)
```

To produce richer output `nb` relies on a **flexible extension API** and a collection of built-in adapters and standalone extensions that allow using other packages to render parts of the notebook:

```go
import (
	"github.com/bevzzz/nb"
	synth "github.com/bevzzz/nb-synth"
	"github.com/bevzzz/nb/extension"
	"github.com/bevzzz/nb/extension/adapter"
	jupyter "github.com/bevzzz/nb/extension/extra/goldmark-jupyter"
	"github.com/robert-nix/ansihtml"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	
)

md := goldmark.New(
	goldmark.WithExtensions(
		jupyter.Attachments(),
		highlighting.Highlighting,
	),
)

c := nb.New(
	nb.WithExtensions(
		jupyter.Goldmark(md),
		synth.Highlighting,
		extension.NewStream(
			adapter.AnsiHtml(ansihtml.ConvertToHTML),
		),
	),
)

err := nb.Convert(os.Stdout, b)
```

The snippet above uses these additional dependencies:

- [`goldmark`](https://github.com/yuin/goldmark) with [`goldmark-jupyter`](./extension/extra/goldmark-jupyter)
- [`chroma`](https://github.com/alecthomas/chroma) with [`nb-synth`](https://github.com/bevzzz/nb-synth)
- [`ansihtml`](https://github.com/robert-nix/ansihtml) with built-in [`adapters.AnsiHtml`](./extension/adapter/ansi.go)

It's a combination of packages that worked really well for me; I encourage you to play around with this [**example CLI**](./example/nbee) to see how it renders different kind of notebooks.

Extending `nb` does not end here. Your project may already use a different Markdown renderer, or require custom handling of certain mime-/cell types, in which case I hope the existing extensions will serve as useful reference implementations.

### Styling the notebook: batteries included 🔋

`nb` comes with the Jupyter's classic light theme, which you can capture by passing a dedicated `CSSWriter` and adding it to the final HTML.

Mind that [the default theme is ~1000 lines long](./render/html/styles/jupyter.css) and might not fit the existing style in a more complex project.  
In that case you probably want to write your own CSS.

<details><summary>Click to expand</summary>

```go
// Write both CSS and notebook's HTML to intermediate destinations
var body, css bytes.Buffer

// Configure your converter
c := nb.New(
  	nb.WithRenderOptions(
		render.WithCellRenderers(
			html.NewRenderer(
				html.WithCSSWriter(&css),
			),
		),
	),
)

err := c.Convert(&body, b)
if err != nil {
	panic(err)
}

// Create the final output
f, _ := os.OpenFile("notebook.html", os.O_RDWR, 0644)
defer f.Close()

f.WriteString("<html><head><style>")
io.Copy(f, &css)
f.WriteString("</style></head>")

f.WriteString("<body>")
io.Copy(f, &body)
f.WriteString("</body></html>")
```

</details>

## Roadmap 🗺

- **v0.4.0**:
  - Built-in pretty-printing for JSON outputs
  - Custom CSS (class prefix / class names).
  I really like the way [`chroma`](https://github.com/alecthomas/chroma/blob/master/formatters/html/html.go) exposes its styling API and I'll try to do something similar.
- Other:
  - I am curious about how `nb`'s performance measures against other popular libraries like [`nbconvert`](https://github.com/jupyter/nbconvert) (Python) and [`quarto`](https://github.com/quarto-dev/quarto-cli) (Javascript), so I want to do some benchmarking later.
  - As of now, I am not planning on adding converters to other formats (LaTeX, PDF, reStructuredText), but I will gladly consider this if there's a need for those.

If you have any other ideas or requests, please feel welcome to add a proposal in a new issue.

## Miscellaneous

### Math

Since Jupyter notebooks are often used for scientific work, you may want to display mathematical notation in your output.  
[MathJax](https://www.mathjax.org) is a powerful tool for that and [adding it to your  HTML header](https://www.mathjax.org/#gettingstarted) is the simplest way to get started.

Notice, that we want to _remove_ `<pre>` from the the list of skipped tags, as default HTML renderer will wrap raw and markdown cells in a `<pre>` tag.

```html
<html>
	<head>
		<script>
		    MathJax = {
		      options: {
		        skipHtmlTags: [ // includes "pre" by default
			        "script",
			        "noscript",
			        "style",
			        "textarea",
			        "code",
			        "annotation",
			        "annotation-xml"
			    ],
		      }
		    };
		</script>
	</head>
</html>
```

MathJax is very configurable and you can read more about that [here](https://docs.mathjax.org/en/latest/options/document.html#document-options).  
You may also find the [official MathJax config](https://nbformat.readthedocs.io/en/latest/markup.html#mathjax-configuration) used in the Jupyter project useful.

## License

This software is released under [the MIT License](https://opensource.org/license/mit/).
