package ace

import "strings"

// ParseSource parses the source and returns the result.
func ParseSource(src *source, opts *Options) (*result, error) {
	// Initialize the options.
	opts = InitializeOptions(opts)

	rslt := newResult(nil, nil, nil)

	base, err := parseBytes(src.base.data, rslt, src, opts, src.base)
	if err != nil {
		return nil, err
	}

	inner, err := parseBytes(src.inner.data, rslt, src, opts, src.inner)
	if err != nil {
		return nil, err
	}

	includes := make(map[string][]element)

	for _, f := range src.includes {
		includes[f.path], err = parseBytes(f.data, rslt, src, opts, f)
		if err != nil {
			return nil, err
		}
	}

	rslt.base = base
	rslt.inner = inner
	rslt.includes = includes

	return rslt, nil
}

// parseBytes parses the byte data and returns the elements.
func parseBytes(data []byte, rslt *result, src *source, opts *Options, f *File) ([]element, error) {
	var elements []element

	lines := strings.Split(formatLF(string(data)), lf)

	i := 0
	l := len(lines)

	// Ignore the last empty line.
	if l > 0 && lines[l-1] == "" {
		l--
	}

	for i < l {
		// Fetch a line.
		ln := newLine(i+1, lines[i], opts, f)
		i++

		// Ignore the empty line.
		if ln.isEmpty() {
			continue
		}

		if ln.isTopIndent() {
			e, err := newElement(ln, rslt, src, nil, opts)
			if err != nil {
				return nil, err
			}

			// Append child elements to the element.
			if err := appendChildren(e, rslt, lines, &i, l, src, opts, f); err != nil {
				return nil, err
			}

			elements = append(elements, e)
		}
	}

	return elements, nil
}

// appendChildren parses the lines and appends the children to the element.
func appendChildren(parent element, rslt *result, lines []string, i *int, l int, src *source, opts *Options, f *File) error {
	for *i < l {
		// Fetch a line.
		ln := newLine(*i+1, lines[*i], opts, f)

		// Check if the line is a child of the parent.
		ok, err := ln.childOf(parent)

		if err != nil {
			return err
		}

		if !ok {
			return nil
		}

		child, err := newElement(ln, rslt, src, parent, opts)
		if err != nil {
			return err
		}

		parent.AppendChild(child)

		*i++

		if child.CanHaveChildren() {
			if err := appendChildren(child, rslt, lines, i, l, src, opts, f); err != nil {
				return err
			}
		}
	}

	return nil
}
