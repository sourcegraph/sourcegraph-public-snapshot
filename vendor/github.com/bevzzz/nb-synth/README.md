# nb-synth

<img src="./assets/synth.png" width="350" height="150" />

`synt`ax `h`ighlighting for code cells and semi-structured output (JSON/XML) with [`chroma`](https://github.com/alecthomas/chroma).

## Installation

```sh
go get github.com/bevzzz/nb-synth
```

## Usage

Adding rich syntax highlighting to your Jupyter notebooks with `synth` is simple and intuitive.  

Start by importing both libraries to your project:

```go
import (
    "github.com/bevzzz/nb"
    synth "github.com/bevzzz/nb-synth"
)
```

Then, extend `nb` converter and convert your Jupyter notebook:

```go
import (
    // synth allows configuring chroma's lexer and formatter
    // with their "native" options, so you will need to import these
    // before adding them to the extension.
    "github.com/alecthomas/chroma"
    chromahtml "github.com/alecthomas/chroma/formatters/html"
)

// Use default configurations.
c := nb.New(
    nb.WithExtensions(
        synth.Highlighting,
    ),
)

// Or, apply options to control lexing, styling, and formatting.
c := nb.New(
    nb.WithExtensions(
        synth.NewHighlighting(
            synth.WithTokenizeOptions(&chroma.TokenizeOptions{
                EnsureLF: true,
            }),
            synth.WithStyle("monokai"),
            synth.WithFormatOptions(
                chromahtml.WithClasses(true),
                chromahtml.WithLineNumbers(true),
            ),
        ),
    ),
)

if err := c.Convert(io.Stdout, b); err != nil {
    panic(err)
}
```

This package draws a lot from the analogous [`goldmark-highlighting`](https://github.com/yuin/goldmark-highlighting), which adds syntax highlighting to [`goldmark`](https://github.com/yuin/goldmark)'s Markdown. If you've had the chance to work with these packages before, `synth`'s structure and APIs will be familiar to you.

## Attributions

Repository cover: <a href="https://www.vecteezy.com/free-vector/synthesizer">Synthesizer Vectors by Vecteezy</a>

## License

This software is released under [the MIT License](https://opensource.org/license/mit/).
