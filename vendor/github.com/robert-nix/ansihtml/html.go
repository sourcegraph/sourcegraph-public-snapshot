package ansihtml

import (
	"fmt"
	"html"
	"io"
	"strconv"
	"strings"
)

type htmlWriter struct {
	w io.Writer

	bold, faint, italic, underline bool
	slowBlink, rapidBlink          bool
	invert, hide, strikeThrough    bool

	fontIndex uint8

	fraktur, doubleUnderline bool

	proportional bool

	fgColor, bgColor, underlineColor sgrColorState

	framed, encircled, overlined bool

	superscript, subscript bool

	lastParamWasReset bool

	inSpan     bool
	useClasses bool
	noStyles   bool

	// prefix all class names with this
	classPrefix string
}

func (w *htmlWriter) Write(b []byte) (int, error) {
	encodedBuf := []byte(html.EscapeString(string(b)))
	n, err := w.w.Write(encodedBuf)
	if n > len(b) {
		n = len(b)
	}
	return n, err
}

type sgrColorState struct {
	// changed => color is not default
	// is24bit => must use style:color(r,g,b), etc.
	changed, is24Bit bool
	// index => 0-15 for standard 4-bit colors,
	//       => 16-255 for 256-color modes
	index   uint8
	r, g, b uint8

	// state variable for handling color parameters
	expectParams uint8
}

const (
	colorParamStateNeedFirstParam uint8 = 1 + iota
	colorParamStateNeed256ColorParam
	colorParamStateNeedRValue
	colorParamStateNeedGValue
	colorParamStateNeedBValue
)

func (s *sgrColorState) handleParam(param int) bool {
	switch s.expectParams {
	case colorParamStateNeedFirstParam:
		switch param {
		case 2:
			s.expectParams = colorParamStateNeedRValue
		case 5:
			s.expectParams = colorParamStateNeed256ColorParam
		default:
			s.expectParams = 0
			return true
		}
	case colorParamStateNeed256ColorParam:
		s.index = uint8(param & 255)
		s.is24Bit = false
		s.changed = true
		s.expectParams = 0
	case colorParamStateNeedRValue:
		s.r = uint8(param)
		s.expectParams = colorParamStateNeedGValue
	case colorParamStateNeedGValue:
		s.g = uint8(param)
		s.expectParams = colorParamStateNeedBValue
	case colorParamStateNeedBValue:
		s.b = uint8(param)
		s.is24Bit = true
		s.changed = true
		s.expectParams = 0
	default:
		return false
	}
	return true
}

var fourBitColors = []byte{
	0x00, 0x00, 0x00, // black
	0x80, 0x00, 0x00, // red
	0x00, 0x80, 0x00, // green
	0x80, 0x80, 0x00, // yellow
	0x00, 0x00, 0x80, // blue
	0x80, 0x00, 0x80, // magenta
	0x00, 0x80, 0x80, // cyan
	0xc0, 0xc0, 0xc0, // white
	0x80, 0x80, 0x80, // bright-black
	0xff, 0x00, 0x00, // bright-red
	0x00, 0xff, 0x00, // bright-green
	0xff, 0xff, 0x00, // bright-yellow
	0x00, 0x00, 0xff, // bright-blue
	0xff, 0x00, 0xff, // bright-magenta
	0x00, 0xff, 0xff, // bright-cyan
	0xff, 0xff, 0xff, // bright-white
}

var cubeColorBytes = []byte{
	0x00, 0x5f, 0x87, 0xaf, 0xd7, 0xff,
}

func (s *sgrColorState) toRGB() (uint8, uint8, uint8) {
	if s.is24Bit {
		return s.r, s.g, s.b
	}
	if s.index < 16 {
		return fourBitColors[s.index*3], fourBitColors[s.index*3+1], fourBitColors[s.index*3+2]
	}
	if s.index >= 16 && s.index <= 231 {
		i := s.index - 16
		bi := i % 6
		i /= 6
		gi := i % 6
		ri := i / 6
		return cubeColorBytes[ri], cubeColorBytes[gi], cubeColorBytes[bi]
	}
	level := (s.index-232)*10 + 8
	return level, level, level
}

const escapeCSIFinalByte byte = '['
const escapeSGRFinalParamByte byte = 'm'

var closeSpan = []byte("</span>")

func (w *htmlWriter) handleEscape(finalByte byte, intermediateBytes, parameterBytes []byte) error {
	// Only CSIs
	if finalByte != escapeCSIFinalByte {
		return nil
	}
	n := len(parameterBytes)
	if n == 0 {
		return nil
	}
	// Only SGRs
	if parameterBytes[n-1] != escapeSGRFinalParamByte {
		return nil
	}

	var param int
	for i := 0; i < n; i++ {
		b := parameterBytes[i]
		if b == ';' {
			w.applyEffect(param)
			param = 0
		} else if b >= '0' && b <= '9' {
			// saturate to LONG_MAX
			if param >= 214748365 || (param == 214748364 && b >= '8') {
				param = 2147483647
			} else {
				param = param*10 + int(b-'0')
			}
		} else if b == 'm' {
			w.applyEffect(param)
			break
		} else {
			// ignore invalidly specified SGRs
			return nil
		}
	}

	err := w.closeSpan()
	if err != nil {
		return err
	}

	if w.lastParamWasReset {
		return nil
	}

	return w.writeSpanOpen()
}

func (w *htmlWriter) closeSpan() error {
	if !w.inSpan {
		return nil
	}
	_, err := w.w.Write(closeSpan)
	if err != nil {
		return err
	}
	w.inSpan = false
	return nil
}

var colorClassNames = []string{
	"black",
	"red",
	"green",
	"yellow",
	"blue",
	"magenta",
	"cyan",
	"white",
	"bright-black",
	"bright-red",
	"bright-green",
	"bright-yellow",
	"bright-blue",
	"bright-magenta",
	"bright-cyan",
	"bright-white",
}

var colorStyles = []string{
	"black",
	"sienna",
	"seagreen",
	"olive",
	"rebeccapurple",
	"darkmagenta",
	"darkturquoise",
	"lightsteelblue",
	"gray",
	"orangered",
	"lawngreen",
	"gold",
	"royalblue",
	"orchid",
	"turquoise",
	"white",
}

func (w *htmlWriter) writeSpanOpen() error {
	var classes []string
	var styles []string
	spanOpen := "<span"
	p := w.classPrefix
	if w.useClasses {
		if w.bold {
			classes = append(classes, p+"bold")
		} else if w.faint {
			classes = append(classes, p+"faint")
		}
		if w.italic {
			classes = append(classes, p+"italic")
		} else if w.fraktur {
			classes = append(classes, p+"fraktur")
		}
		if w.doubleUnderline {
			classes = append(classes, p+"double-underline")
		} else if w.underline {
			classes = append(classes, p+"underline")
		}
		if w.strikeThrough {
			classes = append(classes, p+"strikethrough")
		}
		if w.overlined {
			classes = append(classes, p+"overline")
		}
		if w.slowBlink {
			classes = append(classes, p+"slow-blink")
		} else if w.rapidBlink {
			classes = append(classes, p+"fast-blink")
		}
		if w.invert {
			classes = append(classes, p+"invert")
		}
		if w.hide {
			classes = append(classes, p+"hide")
		}
		if w.fontIndex != 0 {
			classes = append(classes, p+"font-"+strconv.Itoa(int(w.fontIndex)))
		}
		if w.proportional {
			classes = append(classes, p+"proportional")
		}
		if w.superscript {
			classes = append(classes, p+"superscript")
		} else if w.subscript {
			classes = append(classes, p+"subscript")
		}
		if w.fgColor.changed && !w.fgColor.is24Bit && w.fgColor.index <= 15 {
			classes = append(classes, p+"fg-"+colorClassNames[w.fgColor.index])
		}
		if w.bgColor.changed && !w.bgColor.is24Bit && w.bgColor.index <= 15 {
			classes = append(classes, p+"bg-"+colorClassNames[w.bgColor.index])
		}
		if w.underlineColor.changed && !w.underlineColor.is24Bit && w.underlineColor.index <= 15 {
			classes = append(classes, p+"underline-"+colorClassNames[w.underlineColor.index])
		}
	} else {
		if w.bold {
			styles = append(styles, p+"font-weight:bold")
		} else if w.faint {
			styles = append(styles, p+"font-weight:lighter")
		}
		if w.italic {
			styles = append(styles, p+"font-style:italic")
		}
		// text-decoration-line isn't additive with multiple declarations
		var lineStyles []string
		if w.underline {
			lineStyles = append(lineStyles, "underline")
		}
		if w.strikeThrough {
			lineStyles = append(lineStyles, "line-through")
		}
		if w.overlined {
			lineStyles = append(lineStyles, "overline")
		}
		if lineStyles != nil {
			styles = append(styles, "text-decoration-line:"+strings.Join(lineStyles, " "))
		}
		if w.invert {
			styles = append(styles, "filter:invert(100%)")
		}
		if w.hide {
			styles = append(styles, "opacity:0")
		}
		if w.proportional {
			styles = append(styles, "font-family:sans-serif")
		}
		if w.superscript {
			styles = append(styles, "vertical-align:super")
		} else if w.subscript {
			styles = append(styles, "vertical-align:sub")
		}
		if w.fgColor.changed && !w.fgColor.is24Bit && w.fgColor.index <= 15 {
			styles = append(styles, "color:"+colorStyles[w.fgColor.index])
		}
		if w.bgColor.changed && !w.bgColor.is24Bit && w.bgColor.index <= 15 {
			styles = append(styles, "background-color:"+colorStyles[w.bgColor.index])
		}
		if w.underlineColor.changed && !w.underlineColor.is24Bit && w.underlineColor.index <= 15 {
			styles = append(styles, "text-decoration-color:"+colorStyles[w.underlineColor.index])
		}
	}
	if !w.noStyles {
		if w.fgColor.changed && (w.fgColor.is24Bit || w.fgColor.index > 15) {
			r, g, b := w.fgColor.toRGB()
			styles = append(styles, fmt.Sprintf("color:rgb(%d,%d,%d)", r, g, b))
		}
		if w.bgColor.changed && (w.bgColor.is24Bit || w.bgColor.index > 15) {
			r, g, b := w.bgColor.toRGB()
			styles = append(styles, fmt.Sprintf("background-color:rgb(%d,%d,%d)", r, g, b))
		}
		if w.underlineColor.changed && (w.underlineColor.is24Bit || w.underlineColor.index > 15) {
			r, g, b := w.underlineColor.toRGB()
			styles = append(styles, fmt.Sprintf("text-decoration-color:rgb(%d,%d,%d)", r, g, b))
		}
	}

	if len(classes) == 0 && len(styles) == 0 {
		return nil
	}

	if len(classes) > 0 {
		spanOpen += " class=\"" + strings.Join(classes, " ") + "\""
	}
	if len(styles) > 0 {
		spanOpen += " style=\"" + strings.Join(styles, ";") + ";\""
	}
	spanOpen += ">"

	_, err := w.w.Write([]byte(spanOpen))
	if err == nil {
		w.inSpan = true
	}
	return err
}

const (
	sgrEffectReset = iota
	sgrEffectBold
	sgrEffectFaint
	sgrEffectItalic
	sgrEffectUnderline
	sgrEffectSlowBlink
	sgrEffectRapidBlink
	sgrEffectInvert
	sgrEffectHide
	sgrEffectStrikeThrough
	sgrEffectFontDefault
	sgrEffectFontMax = 8 + iota
	sgrEffectFraktur
	sgrEffectDoubleUnderline
	sgrEffectNormalIntensity
	sgrEffectNotItalic
	sgrEffectNotUnderline
	sgrEffectNotBlinking
	sgrEffectProportional
	sgrEffectNotInvert
	sgrEffectNotHide
	sgrEffectNotStrikeThrough
	sgrEffectForegroundColor
	sgrEffectForegroundColorParams = 15 + iota
	sgrEffectForegroundColorDefault
	sgrEffectBackgroundColor
	sgrEffectBackgroundColorParams = 22 + iota
	sgrEffectBackgroundColorDefault
	sgrEffectNotProportional
	sgrEffectFramed
	sgrEffectEncircled
	sgrEffectOverlined
	sgrEffectNotFramedOrEncircled
	sgrEffectNotOverlined
	sgrEffectUnderlineColor        = 58
	sgrEffectUnderlineColorDefault = 59
	sgrEffectSuperscript           = 73
	sgrEffectSubscript             = 74

	sgrEffectForegroundColorBright    = 90
	sgrEffectForegroundColorBrightMax = 97

	sgrEffectBackgroundColorBright    = 100
	sgrEffectBackgroundColorBrightMax = 107
)

func (w *htmlWriter) applyEffect(effect int) {
	w.lastParamWasReset = false
	if w.fgColor.handleParam(effect) {
		return
	}
	if w.bgColor.handleParam(effect) {
		return
	}
	if w.underlineColor.handleParam(effect) {
		return
	}
	switch {
	case effect >= sgrEffectFontDefault && effect < sgrEffectFontMax:
		w.fontIndex = uint8(effect - sgrEffectFontDefault)
	case effect >= sgrEffectForegroundColor && effect < sgrEffectForegroundColorParams:
		w.fgColor.changed = true
		w.fgColor.is24Bit = false
		w.fgColor.index = uint8(effect - sgrEffectForegroundColor)
	case effect >= sgrEffectBackgroundColor && effect < sgrEffectBackgroundColorParams:
		w.bgColor.changed = true
		w.bgColor.is24Bit = false
		w.bgColor.index = uint8(effect - sgrEffectBackgroundColor)
	case effect >= sgrEffectForegroundColorBright && effect <= sgrEffectForegroundColorBrightMax:
		w.fgColor.changed = true
		w.fgColor.is24Bit = false
		w.fgColor.index = 8 + uint8(effect-sgrEffectForegroundColorBright)
	case effect >= sgrEffectBackgroundColorBright && effect <= sgrEffectBackgroundColorBrightMax:
		w.bgColor.changed = true
		w.bgColor.is24Bit = false
		w.bgColor.index = 8 + uint8(effect-sgrEffectBackgroundColorBright)
	}
	switch effect {
	case sgrEffectReset:
		w.lastParamWasReset = true
		w.bold = false
		w.faint = false
		w.italic = false
		w.underline = false
		w.slowBlink = false
		w.rapidBlink = false
		w.invert = false
		w.hide = false
		w.strikeThrough = false
		w.fontIndex = 0
		w.fraktur = false
		w.doubleUnderline = false
		w.proportional = false
		w.fgColor.changed = false
		w.bgColor.changed = false
		w.underlineColor.changed = false
		w.framed = false
		w.encircled = false
		w.overlined = false
		w.superscript = false
		w.subscript = false
	case sgrEffectBold:
		w.bold = true
		w.faint = false
	case sgrEffectFaint:
		w.faint = true
		w.bold = false
	case sgrEffectItalic:
		w.italic = true
		w.fraktur = false
	case sgrEffectUnderline:
		w.underline = true
		w.doubleUnderline = false
	case sgrEffectSlowBlink:
		w.slowBlink = true
		w.rapidBlink = false
	case sgrEffectRapidBlink:
		w.rapidBlink = true
		w.slowBlink = false
	case sgrEffectInvert:
		w.invert = true
	case sgrEffectHide:
		w.hide = true
	case sgrEffectStrikeThrough:
		w.strikeThrough = true
	case sgrEffectFraktur:
		w.fraktur = true
		w.italic = false
	case sgrEffectDoubleUnderline:
		w.doubleUnderline = true
		w.underline = false
	case sgrEffectNormalIntensity:
		w.bold = false
		w.faint = false
	case sgrEffectNotItalic:
		w.italic = false
		w.fraktur = false
	case sgrEffectNotUnderline:
		w.underline = false
		w.doubleUnderline = false
	case sgrEffectNotBlinking:
		w.slowBlink = false
		w.rapidBlink = false
	case sgrEffectProportional:
		w.proportional = true
	case sgrEffectNotInvert:
		w.invert = false
	case sgrEffectNotHide:
		w.hide = false
	case sgrEffectNotStrikeThrough:
		w.strikeThrough = false
	case sgrEffectForegroundColorParams:
		w.fgColor.expectParams = colorParamStateNeedFirstParam
	case sgrEffectForegroundColorDefault:
		w.fgColor.changed = false
	case sgrEffectBackgroundColorParams:
		w.bgColor.expectParams = colorParamStateNeedFirstParam
	case sgrEffectBackgroundColorDefault:
		w.bgColor.changed = false
	case sgrEffectNotProportional:
		w.proportional = false
	case sgrEffectFramed:
		w.framed = true
		w.encircled = false
	case sgrEffectEncircled:
		w.encircled = true
		w.framed = false
	case sgrEffectOverlined:
		w.overlined = true
	case sgrEffectNotFramedOrEncircled:
		w.framed = false
		w.encircled = false
	case sgrEffectNotOverlined:
		w.overlined = false
	case sgrEffectUnderlineColor:
		w.underlineColor.expectParams = colorParamStateNeedFirstParam
	case sgrEffectUnderlineColorDefault:
		w.underlineColor.changed = false
	case sgrEffectSuperscript:
		w.superscript = true
		w.subscript = false
	case sgrEffectSubscript:
		w.subscript = true
		w.superscript = false
	}
}
