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
	args   []interface{}
}

// Line creates a new FancyLine without a format string.
func Line(emoji string, style Style, s string) FancyLine {
	return FancyLine{
		emoji:  emoji,
		style:  style,
		format: "%s",
		args:   []interface{}{s},
	}
}

// Line creates a new FancyLine with a format string. As with Writer, the
// arguments may include Style instances with the %s specifier.
func Linef(emoji string, style Style, format string, a ...interface{}) FancyLine {
	args := make([]interface{}, len(a))
	for i := range a {
		args[i] = a[i]
	}

	return FancyLine{
		emoji:  emoji,
		style:  style,
		format: format,
		args:   args,
	}
}

func (ol FancyLine) write(w io.Writer, caps capabilities) {
	if ol.emoji != "" {
		fmt.Fprint(w, ol.emoji+" ")
	}

	fmt.Fprintf(w, "%s"+ol.format+"%s", caps.formatArgs(append(append([]interface{}{ol.style}, ol.args...), StyleReset))...)
	w.Write([]byte("\n"))
}
