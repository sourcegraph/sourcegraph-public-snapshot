package ace

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// Tag names
const (
	tagNameDiv = "div"
)

// Attribute names
const (
	attributeNameID = "id"
)

// htmlAttribute represents an HTML attribute.
type htmlAttribute struct {
	key   string
	value string
}

// htmlTag represents an HTML tag.
type htmlTag struct {
	elementBase
	tagName          string
	id               string
	classes          []string
	containPlainText bool
	insertBr         bool
	attributes       []htmlAttribute
	textValue        string
}

// WriteTo writes data to w.
func (e *htmlTag) WriteTo(w io.Writer) (int64, error) {
	var bf bytes.Buffer

	// Write an open tag.
	bf.WriteString(lt)
	bf.WriteString(e.tagName)
	// Write an id.
	if e.id != "" {
		bf.WriteString(space)
		bf.WriteString(attributeNameID)
		bf.WriteString(equal)
		bf.WriteString(doubleQuote)
		bf.WriteString(e.id)
		bf.WriteString(doubleQuote)
	}
	// Write classes.
	if len(e.classes) > 0 {
		bf.WriteString(space)
		bf.WriteString(e.opts.AttributeNameClass)
		bf.WriteString(equal)
		bf.WriteString(doubleQuote)
		for i, class := range e.classes {
			if i > 0 {
				bf.WriteString(space)
			}
			bf.WriteString(class)
		}
		bf.WriteString(doubleQuote)
	}
	// Write attributes.
	if len(e.attributes) > 0 {

		for _, a := range e.attributes {
			bf.WriteString(space)
			bf.WriteString(a.key)
			if a.value != "" {
				bf.WriteString(equal)
				bf.WriteString(doubleQuote)
				bf.WriteString(a.value)
				bf.WriteString(doubleQuote)
			}
		}
	}
	bf.WriteString(gt)

	// Write a text value
	if e.textValue != "" {
		bf.WriteString(e.textValue)
	}

	if e.containPlainText {
		bf.WriteString(lf)
	}

	// Write children's HTML.
	if i, err := e.writeChildren(&bf); err != nil {
		return i, err
	}

	// Write a close tag.
	if !e.noCloseTag() {
		bf.WriteString(lt)
		bf.WriteString(slash)
		bf.WriteString(e.tagName)
		bf.WriteString(gt)
	}

	// Write the buffer.
	i, err := w.Write(bf.Bytes())

	return int64(i), err
}

// ContainPlainText returns the HTML tag's containPlainText field.
func (e *htmlTag) ContainPlainText() bool {
	return e.containPlainText
}

// InsertBr returns true if the br tag is inserted to the line.
func (e *htmlTag) InsertBr() bool {
	return e.insertBr
}

// setAttributes parses the tokens and set attributes to the element.
func (e *htmlTag) setAttributes() error {
	parsedTokens := e.parseTokens()

	var i int
	var token string
	var setTextValue bool

	// Set attributes to the element.
	for i, token = range parsedTokens {
		kv := strings.Split(token, equal)

		if len(kv) < 2 {
			setTextValue = true
			break
		}

		k := kv[0]
		v := strings.Join(kv[1:], equal)

		// Remove the prefix and suffix of the double quotes.
		if len(v) > 1 && strings.HasPrefix(v, doubleQuote) && strings.HasSuffix(v, doubleQuote) {
			v = v[1 : len(v)-1]
		}

		switch k {
		case attributeNameID:
			if e.id != "" {
				return fmt.Errorf("multiple IDs are specified [file: %s][line: %d]", e.ln.fileName(), e.ln.no)
			}
			e.id = v
		case e.opts.AttributeNameClass:
			e.classes = append(e.classes, strings.Split(v, space)...)
		default:
			e.attributes = append(e.attributes, htmlAttribute{k, v})
		}
	}

	// Set a text value to the element.
	if setTextValue {
		e.textValue = strings.Join(parsedTokens[i:], space)
	}

	return nil
}

// noCloseTag returns true is the HTML tag has no close tag.
func (e *htmlTag) noCloseTag() bool {
	for _, name := range e.opts.NoCloseTagNames {
		if e.tagName == name {
			return true
		}
	}

	return false
}

// newHTMLTag creates and returns an HTML tag.
func newHTMLTag(ln *line, rslt *result, src *source, parent element, opts *Options) (*htmlTag, error) {
	if len(ln.tokens) < 1 {
		return nil, fmt.Errorf("an HTML tag is not specified [file: %s][line: %d]", ln.fileName(), ln.no)
	}

	s := ln.tokens[0]

	tagName := extractTagName(s)

	id, err := extractID(s, ln)
	if err != nil {
		return nil, err
	}

	classes := extractClasses(s)

	e := &htmlTag{
		elementBase:      newElementBase(ln, rslt, src, parent, opts),
		tagName:          tagName,
		id:               id,
		classes:          classes,
		containPlainText: strings.HasSuffix(s, dot),
		insertBr:         strings.HasSuffix(s, doubleDot),
		attributes:       make([]htmlAttribute, 0, 2),
	}

	if err := e.setAttributes(); err != nil {
		return nil, err
	}

	return e, nil
}

// extractTag extracts and returns a tag.
func extractTagName(s string) string {
	tagName := strings.Split(strings.Split(s, sharp)[0], dot)[0]

	if tagName == "" {
		tagName = tagNameDiv
	}

	return tagName
}

// extractID extracts and returns an ID.
func extractID(s string, ln *line) (string, error) {
	tokens := strings.Split(s, sharp)

	l := len(tokens)

	if l < 2 {
		return "", nil
	}

	if l > 2 {
		return "", fmt.Errorf("multiple IDs are specified [file: %s][line: %d]", ln.fileName(), ln.no)
	}

	return strings.Split(tokens[1], dot)[0], nil
}

// extractClasses extracts and returns classes.
func extractClasses(s string) []string {
	var classes []string

	for i, token := range strings.Split(s, dot) {
		if i == 0 {
			continue
		}

		class := strings.Split(token, sharp)[0]

		if class == "" {
			continue
		}

		classes = append(classes, class)
	}

	return classes
}

// parseTokens parses the tokens and return them
func (e *htmlTag) parseTokens() []string {
	var inQuote bool
	var inDelim bool
	var tokens []string
	var token string

	str := strings.Join(e.ln.tokens[1:], space)
	for _, chr := range str {
		switch c := string(chr); c {
		case space:
			if inQuote || inDelim {
				token += c
			} else {
				tokens = append(tokens, token)
				token = ""
			}
		case doubleQuote:
			if !inDelim {
				if inQuote {
					inQuote = false
				} else {
					inQuote = true
				}
			}
			token += c
		default:
			token += c
			if inDelim {
				if strings.HasSuffix(token, e.opts.DelimRight) {
					inDelim = false
				}
			} else {
				if strings.HasSuffix(token, e.opts.DelimLeft) {
					inDelim = true
				}
			}
		}
	}
	if len(token) > 0 {
		tokens = append(tokens, token)
	}
	return tokens
}
