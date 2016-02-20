// Functions to parse quote-like elements.

package mmark

import "bytes"

// TODO(miek): this can be refactered

// returns notequote prefix length
func (p *parser) notePrefix(data []byte) int {
	i := 0
	for i < 3 && data[i] == ' ' {
		i++
	}
	if data[i] == 'N' && data[i+1] == '>' {
		if data[i+2] == ' ' {
			return i + 3
		}
		return i + 2
	}
	return 0
}

// parse an note fragment
func (p *parser) note(out *bytes.Buffer, data []byte) int {
	var raw bytes.Buffer
	beg, end := 0, 0
	for beg < len(data) {
		end = beg
		for data[end] != '\n' {
			end++
		}
		end++

		if pre := p.notePrefix(data[beg:]); pre > 0 {
			// skip the prefix
			beg += pre
		} else if p.isEmpty(data[beg:]) > 0 &&
			(end >= len(data) ||
				(p.notePrefix(data[end:]) == 0 && p.isEmpty(data[end:]) == 0)) {
			break
		}
		raw.Write(data[beg:end])
		beg = end
	}

	var cooked bytes.Buffer
	p.block(&cooked, raw.Bytes())

	p.r.SetInlineAttr(p.ial)
	p.ial = nil

	p.r.Note(out, cooked.Bytes())
	return end
}

// returns asidequote prefix length
func (p *parser) asidePrefix(data []byte) int {
	i := 0
	for i < 3 && data[i] == ' ' {
		i++
	}
	if data[i] == 'A' && data[i+1] == '>' {
		if data[i+2] == ' ' {
			return i + 3
		}
		return i + 2
	}
	return 0
}

// parse an aside fragment
func (p *parser) aside(out *bytes.Buffer, data []byte) int {
	var raw bytes.Buffer
	beg, end := 0, 0
	for beg < len(data) {
		end = beg
		for data[end] != '\n' {
			end++
		}
		end++

		if pre := p.asidePrefix(data[beg:]); pre > 0 {
			// skip the prefix
			beg += pre
		} else if p.isEmpty(data[beg:]) > 0 &&
			(end >= len(data) ||
				(p.asidePrefix(data[end:]) == 0 && p.isEmpty(data[end:]) == 0)) {
			break
		}
		raw.Write(data[beg:end])
		beg = end
	}

	var cooked bytes.Buffer
	p.block(&cooked, raw.Bytes())

	p.r.SetInlineAttr(p.ial)
	p.ial = nil

	p.r.Aside(out, cooked.Bytes())
	return end
}

// returns blockquote prefix length
func (p *parser) quotePrefix(data []byte) int {
	i := 0
	for i < 3 && data[i] == ' ' {
		i++
	}
	if data[i] == '>' {
		if data[i+1] == ' ' {
			return i + 2
		}
		return i + 1
	}
	return 0
}

// parse a blockquote fragment
func (p *parser) quote(out *bytes.Buffer, data []byte) int {
	var raw bytes.Buffer
	beg, end := 0, 0
	for beg < len(data) {
		end = beg
		for data[end] != '\n' {
			end++
		}
		end++

		if pre := p.quotePrefix(data[beg:]); pre > 0 {
			// skip the prefix
			beg += pre
		} else if bytes.HasPrefix(data[beg:], []byte("Quote: ")) {
			break
		} else if p.isEmpty(data[beg:]) > 0 &&
			(end >= len(data) ||
				(p.quotePrefix(data[end:]) == 0 && p.isEmpty(data[end:]) == 0)) {
			// blockquote ends with at least one blank line
			// followed by something without a blockquote prefix
			break
		}

		// this line is part of the blockquote
		raw.Write(data[beg:end])
		beg = end
	}
	var attribution bytes.Buffer
	line := beg
	j := beg
	if bytes.HasPrefix(data[j:], []byte("Quote: ")) {
		for line < len(data) {
			j++
			// find the end of this line
			for data[j-1] != '\n' {
				j++
			}
			if p.isEmpty(data[line:j]) > 0 {
				break
			}
			line = j
		}
		p.inline(&attribution, data[beg+7:j-1]) // +7 for 'Quote: '
	}
	ials := p.ial
	p.ial = nil

	// This set a new level of attributes, we could possible override what
	// we have gotten above. TODO(miek): this might need to happen in more
	// places.
	var cooked bytes.Buffer
	p.block(&cooked, raw.Bytes())

	p.r.SetInlineAttr(ials)

	p.r.BlockQuote(out, cooked.Bytes(), attribution.Bytes())
	return j
}

// returns figurequote prefix length
func (p *parser) figurePrefix(data []byte) int {
	i := 0
	for i < 3 && data[i] == ' ' {
		i++
	}
	if data[i] == 'F' && data[i+1] == '>' {
		if data[i+2] == ' ' {
			return i + 3
		}
		return i + 2
	}
	return 0
}

// parse a figurequote fragment
func (p *parser) figure(out *bytes.Buffer, data []byte) int {
	var raw bytes.Buffer
	beg, end := 0, 0
	for beg < len(data) {
		end = beg
		for data[end] != '\n' {
			end++
		}
		end++

		if pre := p.figurePrefix(data[beg:]); pre > 0 {
			// skip the prefix
			beg += pre
		} else if bytes.HasPrefix(data[beg:], []byte("Figure: ")) {
			break
		} else if p.isEmpty(data[beg:]) > 0 &&
			(end >= len(data) ||
				(p.figurePrefix(data[end:]) == 0 && p.isEmpty(data[end:]) == 0)) {
			// figurequote ends with at least one blank line
			// followed by something without a figurequote prefix
			break
		}

		// this line is part of the figurequote
		raw.Write(data[beg:end])
		beg = end
	}
	var caption bytes.Buffer
	line := beg
	j := beg
	// this one must start on j
	if bytes.HasPrefix(data[j:], []byte("Figure: ")) {
		for line < len(data) {
			j++
			// find the end of this line
			for data[j-1] != '\n' {
				j++
			}
			if p.isEmpty(data[line:j]) > 0 {
				break
			}
			line = j
		}
		p.inline(&caption, data[beg+8:j-1]) // +8 for 'Figure: '
	}

	p.insideFigure = true
	var cooked bytes.Buffer
	p.block(&cooked, raw.Bytes())
	p.insideFigure = false

	p.r.SetInlineAttr(p.ial)
	p.ial = nil

	p.r.Figure(out, cooked.Bytes(), caption.Bytes())
	return j
}
