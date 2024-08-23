package html2text

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
)

// Line break constants
// Deprecated: Please use HTML2TextWithOptions(text, WithUnixLineBreak())
const (
	WIN_LBR  = "\r\n"
	UNIX_LBR = "\n"
)

var legacyLBR = WIN_LBR
var badTagnamesRE = regexp.MustCompile(`^(head|script|style|a)($|\s+)`)
var linkTagRE = regexp.MustCompile(`a.*href=('([^']*?)'|"([^"]*?)")`)
var badLinkHrefRE = regexp.MustCompile(`javascript:`)
var headersRE = regexp.MustCompile(`^(\/)?h[1-6]`)
var numericEntityRE = regexp.MustCompile(`(?i)^#(x?[a-f0-9]+)$`)

type options struct {
	lbr            string
	linksInnerText bool
}

func newOptions() *options {
	// apply defaults
	return &options{
		lbr: WIN_LBR,
	}
}

// Option is a functional option
type Option func(*options)

// WithUnixLineBreaks instructs the converter to use unix line breaks ("\n" instead of "\r\n" default)
func WithUnixLineBreaks() Option {
	return func(o *options) {
		o.lbr = UNIX_LBR
	}
}

// WithLinksInnerText instructs the converter to retain link tag inner text and append href URLs in angle brackets after the text
// Example: click news <http://bit.ly/2n4wXRs>
func WithLinksInnerText() Option {
	return func(o *options) {
		o.linksInnerText = true
	}
}

func parseHTMLEntity(entName string) (string, bool) {
	if r, ok := entity[entName]; ok {
		return string(r), true
	}

	if match := numericEntityRE.FindStringSubmatch(entName); len(match) == 2 {
		var (
			err    error
			n      int64
			digits = match[1]
		)

		if digits != "" && (digits[0] == 'x' || digits[0] == 'X') {
			n, err = strconv.ParseInt(digits[1:], 16, 64)
		} else {
			n, err = strconv.ParseInt(digits, 10, 64)
		}

		if err == nil && (n == 9 || n == 10 || n == 13 || n > 31) {
			return string(rune(n)), true
		}
	}

	return "", false
}

// SetUnixLbr with argument true sets Unix-style line-breaks in output ("\n")
// with argument false sets Windows-style line-breaks in output ("\r\n", the default)
// Deprecated: Please use HTML2TextWithOptions(text, WithUnixLineBreak())
func SetUnixLbr(b bool) {
	if b {
		legacyLBR = UNIX_LBR
	} else {
		legacyLBR = WIN_LBR
	}
}

// HTMLEntitiesToText decodes HTML entities inside a provided
// string and returns decoded text
func HTMLEntitiesToText(htmlEntsText string) string {
	outBuf := bytes.NewBufferString("")
	inEnt := false

	for i, r := range htmlEntsText {
		switch {
		case r == ';' && inEnt:
			inEnt = false
			continue

		case r == '&': //possible html entity
			entName := ""
			isEnt := false

			// parse the entity name - max 10 chars
			chars := 0
			for _, er := range htmlEntsText[i+1:] {
				if er == ';' {
					isEnt = true
					break
				} else {
					entName += string(er)
				}

				chars++
				if chars == 10 {
					break
				}
			}

			if isEnt {
				if ent, isEnt := parseHTMLEntity(entName); isEnt {
					outBuf.WriteString(ent)
					inEnt = true
					continue
				}
			}
		}

		if !inEnt {
			outBuf.WriteRune(r)
		}
	}

	return outBuf.String()
}

func writeSpace(outBuf *bytes.Buffer) {
	bts := outBuf.Bytes()
	if len(bts) > 0 && bts[len(bts)-1] != ' ' {
		outBuf.WriteString(" ")
	}
}

// HTML2Text converts html into a text form
func HTML2Text(html string) string {
	var opts []Option
	if legacyLBR == UNIX_LBR {
		opts = append(opts, WithUnixLineBreaks())
	}
	return HTML2TextWithOptions(html, opts...)
}

// HTML2TextWithOptions converts html into a text form with additional options
func HTML2TextWithOptions(html string, reqOpts ...Option) string {
	opts := newOptions()
	for _, opt := range reqOpts {
		opt(opts)
	}

	inLen := len(html)
	tagStart := 0
	inEnt := false
	badTagStackDepth := 0 // if == 1 it means we are inside <head>...</head>
	shouldOutput := true
	// maintain a stack of <a> tag href links and output it after the tag's inner text (for opts.linksInnerText only)
	hrefs := []string{}
	// new line cannot be printed at the beginning or
	// for <p> after a new line created by previous <p></p>
	canPrintNewline := false

	outBuf := bytes.NewBufferString("")

	for i, r := range html {
		if inLen > 0 && i == inLen-1 {
			// prevent new line at the end of the document
			canPrintNewline = false
		}

		switch {
		// skip new lines and spaces adding a single space if not there yet
		case r <= 0xD, r == 0x85, r == 0x2028, r == 0x2029, // new lines
			r == ' ', r >= 0x2008 && r <= 0x200B: // spaces
			if shouldOutput && badTagStackDepth == 0 && !inEnt {
				//outBuf.WriteString(fmt.Sprintf("{DBG r:%c, inEnt:%t, tag:%s}", r, inEnt, html[tagStart:i]))
				writeSpace(outBuf)
			}
			continue

		case r == ';' && inEnt: // end of html entity
			inEnt = false
			continue

		case r == '&' && shouldOutput: // possible html entity
			entName := ""
			isEnt := false

			// parse the entity name - max 10 chars
			chars := 0
			for _, er := range html[i+1:] {
				if er == ';' {
					isEnt = true
					break
				} else {
					entName += string(er)
				}

				chars++
				if chars == 10 {
					break
				}
			}

			if isEnt {
				if ent, isEnt := parseHTMLEntity(entName); isEnt {
					outBuf.WriteString(ent)
					inEnt = true
					continue
				}
			}

		case r == '<': // start of a tag
			tagStart = i + 1
			shouldOutput = false
			continue

		case r == '>': // end of a tag
			shouldOutput = true
			tag := html[tagStart:i]
			tagNameLowercase := strings.ToLower(tag)

			if tagNameLowercase == "/ul" {
				outBuf.WriteString(opts.lbr)
			} else if tagNameLowercase == "li" || tagNameLowercase == "li/" {
				outBuf.WriteString(opts.lbr)
			} else if headersRE.MatchString(tagNameLowercase) {
				if canPrintNewline {
					outBuf.WriteString(opts.lbr + opts.lbr)
				}
				canPrintNewline = false
			} else if tagNameLowercase == "br" || tagNameLowercase == "br/" {
				// new line
				outBuf.WriteString(opts.lbr)
			} else if tagNameLowercase == "p" || tagNameLowercase == "/p" {
				if canPrintNewline {
					outBuf.WriteString(opts.lbr + opts.lbr)
				}
				canPrintNewline = false
			} else if opts.linksInnerText && tagNameLowercase == "/a" {
				// end of link
				// links can be empty can happen if the link matches the badLinkHrefRE
				if len(hrefs) > 0 {
					outBuf.WriteString(" <")
					outBuf.WriteString(HTMLEntitiesToText(hrefs[0]))
					outBuf.WriteString(">")
					hrefs = hrefs[1:]
				}
			} else if opts.linksInnerText && linkTagRE.MatchString(tagNameLowercase) {
				// parse link href
				// add special handling for a tags
				m := linkTagRE.FindStringSubmatch(tag)
				if len(m) == 4 {
					link := m[2]
					if len(link) == 0 {
						link = m[3]
					}

					if opts.linksInnerText && !badLinkHrefRE.MatchString(link) {
						hrefs = append(hrefs, link)
					}
				}
			} else if badTagnamesRE.MatchString(tagNameLowercase) {
				// unwanted block
				badTagStackDepth++

				// if link inner text preservation is not enabled
				// and the current tag is a link tag, parse its href and output that
				if !opts.linksInnerText {
					// parse link href
					m := linkTagRE.FindStringSubmatch(tag)
					if len(m) == 4 {
						link := m[2]
						if len(link) == 0 {
							link = m[3]
						}

						if !badLinkHrefRE.MatchString(link) {
							outBuf.WriteString(HTMLEntitiesToText(link))
						}
					}
				}
			} else if len(tagNameLowercase) > 0 && tagNameLowercase[0] == '/' &&
				badTagnamesRE.MatchString(tagNameLowercase[1:]) {
				// end of unwanted block
				badTagStackDepth--
			}
			continue

		} // switch end

		if shouldOutput && badTagStackDepth == 0 && !inEnt {
			canPrintNewline = true
			outBuf.WriteRune(r)
		}
	}

	return outBuf.String()
}
