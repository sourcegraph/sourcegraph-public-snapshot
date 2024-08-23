// Package ansihtml parses text formatted with ANSI escape sequences and
// outputs text suitable for display in a HTML <pre> tag.
//
// Text effects are encoded as <span> tags with style attributes or various
// classes.
package ansihtml

import "bytes"

// ConvertToHTML converts ansiBytes to HTML where ANSI escape sequences have
// been translated to <span> tags with appropriate style attributes.
func ConvertToHTML(ansiBytes []byte) []byte {
	return convertToHTML(ansiBytes, "", false, false)
}

// ConvertToHTMLWithClasses converts ansiBytes to HTML where ANSI escape
// sequences have been translated to <span> tags with appropriate classes set.
//
// classPrefix will be prefixed to the standard class names in the output.
//
// If noStyles is true, no style tags will be emitted for 256-color and 24-bit
// color text; instead, these color sequences will have no effect.
//
// A span in the output may have any combination of these classes:
//  * 'bold' or 'faint'
//  * 'italic' or 'fraktur'
//  * 'double-underline' or 'underline'
//  * 'strikethrough'
//  * 'overline'
//  * 'slow-blink' or 'fast-blink'
//  * 'invert'
//  * 'hide'
//  * one of 'font-{n}' where n is between 1 and 9
//  * 'proportional'
//  * 'superscript' or 'subscript'
//  * 'fg-{color}', 'bg-{color}', and 'underline-{color}' where color is one of
//   * black
//   * red
//   * green
//   * yellow
//   * blue
//   * magenta
//   * cyan
//   * white
//   * bright-black
//   * bright-red
//   * bright-green
//   * bright-yellow
//   * bright-blue
//   * bright-magenta
//   * bright-cyan
//   * bright-white
func ConvertToHTMLWithClasses(ansiBytes []byte, classPrefix string, noStyles bool) []byte {
	return convertToHTML(ansiBytes, classPrefix, true, noStyles)
}

func convertToHTML(ansiBytes []byte, classPrefix string, useClasses, noStyles bool) []byte {
	rd := bytes.NewBuffer(ansiBytes)
	output := new(bytes.Buffer)
	w := &htmlWriter{
		w:           output,
		useClasses:  useClasses,
		noStyles:    noStyles,
		classPrefix: classPrefix,
	}
	p := NewParser(rd, w)
	err := p.Parse(w.handleEscape)
	w.closeSpan()
	// err must be nil since the underlying readers and writers
	// cannot return errors
	if err != nil {
		return nil
	}
	return output.Bytes()
}
