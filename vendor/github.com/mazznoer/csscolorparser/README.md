# Golang CSS Color Parser Library

[![PkgGoDev](https://pkg.go.dev/badge/github.com/mazznoer/csscolorparser)](https://pkg.go.dev/github.com/mazznoer/csscolorparser)
[![Build Status](https://travis-ci.org/mazznoer/csscolorparser.svg?branch=master)](https://travis-ci.org/mazznoer/csscolorparser)
[![Build Status](https://github.com/mazznoer/csscolorparser/workflows/Go/badge.svg)](https://github.com/mazznoer/csscolorparser/actions)
[![go report](https://goreportcard.com/badge/github.com/mazznoer/csscolorparser)](https://goreportcard.com/report/github.com/mazznoer/csscolorparser)
[![codecov](https://codecov.io/gh/mazznoer/csscolorparser/branch/master/graph/badge.svg)](https://codecov.io/gh/mazznoer/csscolorparser)
[![Lines of Code](https://tokei.rs/b1/github/mazznoer/csscolorparser?category=code)](https://github.com/mazznoer/csscolorparser)

[Go](https://www.golang.org/) library for parsing CSS color string as defined in the W3C's [CSS Color Module Level 4](https://www.w3.org/TR/css-color-4/).

## Supported Color Format

* [Named colors](https://www.w3.org/TR/css-color-4/#named-colors)
* RGB hexadecimal (with and without `#` prefix)
     + Short format `#rgb`
     + Short format with alpha `#rgba`
     + Long format `#rrggbb`
     + Long format with alpha `#rrggbbaa`
* `rgb()` and `rgba()`
* `hsl()` and `hsla()`
* `hwb()`
* `hwba()`, `hsv()`, `hsva()` - not in CSS standard.

Not yet supported: `lab()`, `lch()`.

### Example Color Format

```
transparent
lime
#0f0
#0f0f
#00ff00
#00ff00ff
rgb(0,255,0)
rgb(0% 100% 0%)
rgb(0 255 0 / 100%)
rgba(0,255,0,1)
hsl(120,100%,50%)
hsl(120deg 100% 50%)
hsl(-240 100% 50%)
hsl(-240deg 100% 50%)
hsl(0.3333turn 100% 50%)
hsl(133.333grad 100% 50%)
hsl(2.0944rad 100% 50%)
hsla(120,100%,50%,100%)
hwb(120 0% 0%)
hwb(480deg 0% 0% / 100%)
hsv(120,100%,100%)
hsv(120deg 100% 100% / 100%)
```

## Usage Examples

```go
import "github.com/mazznoer/csscolorparser"
```

```go
c, err := csscolorparser.Parse("gold")

if err != nil {
	panic(err)
}

fmt.Printf("R:%.3f, G:%.3f, B:%.3f, A:%.3f", c.R, c.G, c.B, c.A) // R:1.000, G:0.843, B:0.000, A:1.000
fmt.Println(c.RGBA255())   // 255 215 0 255
fmt.Println(c.HexString()) // #ffd700
fmt.Println(c.RGBString()) // rgb(255,215,0)
```

## Try It Online

* [Playground 1](https://play.golang.org/p/8KMIc1TLQB0)
* [Playground 2](https://play.golang.org/p/7kb62KSARwa)

## Similar Projects

* [csscolorparser](https://github.com/mazznoer/csscolorparser-rs) (Rust)
* [csscolorparser](https://github.com/deanm/css-color-parser-js) (Javascript)

