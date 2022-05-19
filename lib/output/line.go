package output

import (
	"fmt"
	"io"
)

// FancyLine is a formatted output line with an optional emoji and style.
type FancyLine struct {
	emoji  string
	style  Style
	format string
	args   []any

	// Prefix can be set to prepend some content to this fancy line.
	Prefix string
}

// Line creates a new FancyLine without a format string.
func Line(emoji string, style Style, s string) FancyLine {
	return FancyLine{
		emoji:  emoji,
		style:  style,
		format: "%s",
		args:   []any{s},
	}
}

// Line creates a new FancyLine with a format string. As with Writer, the
// arguments may include Style instances with the %s specifier.
func Linef(emoji string, style Style, format string, a ...any) FancyLine {
	return FancyLine{
		emoji:  emoji,
		style:  style,
		format: format,
		args:   a,
	}
}

// Emoji creates a new FancyLine with an emoji prefix.
func Emoji(emoji string, s string) FancyLine {
	return Line(emoji, nil, s)
}

// Emoji creates a new FancyLine with an emoji prefix and style.
func Emojif(emoji string, s string, a ...any) FancyLine {
	return Linef(emoji, nil, s, a...)
}

// Styled creates a new FancyLine with style.
func Styled(style Style, s string) FancyLine {
	return Line("", style, s)
}

// Styledf creates a new FancyLine with style and format string.
func Styledf(style Style, s string, a ...any) FancyLine {
	return Linef("", style, s, a...)
}

func (fl FancyLine) write(w io.Writer, caps capabilities) {
	if fl.Prefix != "" {
		fmt.Fprint(w, fl.Prefix+" ")
	}
	if fl.emoji != "" {
		fmt.Fprint(w, fl.emoji+" ")
	}

	fmt.Fprintf(w, "%s"+fl.format+"%s", caps.formatArgs(append(append([]any{fl.style}, fl.args...), StyleReset))...)
	_, _ = w.Write([]byte("\n"))
}
