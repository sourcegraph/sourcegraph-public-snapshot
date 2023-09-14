package codeowners

import (
	"bufio"
	"io"
	"net/mail"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Parse parses CODEOWNERS file given as a Reader and returns the proto
// representation of all rules within. The rules are in the same order
// as in the file, since this matters for evaluation.
func Parse(codeownersFile io.Reader) (*codeownerspb.File, error) {
	scanner := bufio.NewScanner(codeownersFile)
	var rs []*codeownerspb.Rule
	p := new(parsing)
	lineNumber := int32(0)
	for scanner.Scan() {
		p.nextLine(scanner.Text())
		lineNumber++
		if p.isBlank() {
			continue
		}
		if p.matchSection() {
			continue
		}
		pattern, owners, ok := p.matchRule()
		if !ok {
			return nil, errors.Errorf("failed to match rule: %s", p.line)
		}
		// Need to handle this error once, codeownerspb.File supports
		// error metadata.
		r := codeownerspb.Rule{
			Pattern: unescape(pattern),
			// Section names are case-insensitive, so we lowercase it.
			SectionName: strings.TrimSpace(strings.ToLower(p.section)),
			LineNumber:  lineNumber,
		}
		for _, ownerText := range owners {
			o := ParseOwner(ownerText)
			r.Owner = append(r.Owner, o)
		}
		rs = append(rs, &r)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &codeownerspb.File{Rule: rs}, nil
}

func ParseOwner(ownerText string) *codeownerspb.Owner {
	var o codeownerspb.Owner
	if strings.HasPrefix(ownerText, "@") {
		o.Handle = strings.TrimPrefix(ownerText, "@")
	} else if a, err := mail.ParseAddress(ownerText); err == nil {
		o.Email = a.Address
	} else {
		o.Handle = ownerText
	}
	return &o
}

// parsing implements matching and parsing primitives for CODEOWNERS files
// as well as keeps track of internal state as a file is being parsed.
type parsing struct {
	// line is the current line being parsed. CODEOWNERS files are built
	// in such a way that for syntactic purposes, every line can be considered
	// in isolation.
	line string
	// The most recently defined section, or "" if none.
	section string
}

// nextLine advances parsing to focus on the next line.
func (p *parsing) nextLine(line string) {
	p.line = line
}

// rulePattern is expected to match a rule line like:
// `cmd/**/docs/index.md @readme-owners owner@example.com`.
//
//	^^^^^^^^^^^^^^^^^^^^ ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
//
// The first capturing   The second capturing group
// group extracts        extracts all the owners
// the file pattern.     separated by whitespace.
//
// The first capturing group supports escaping a whitespace with a `\`,
// so that the space is not treated as a separator between the pattern
// and owners.
var rulePattern = lazyregexp.New(`^\s*((?:\\.|\S)+)((?:\s+\S+)*)\s*$`)

// matchRule tries to extract a codeowners rule from the current line
// and return the file pattern and one or more owners.
// Match is indicated by the third return value being true.
//
// Note: Need to check if a line matches a section using `matchSection`
// before matching a rule with this method, as `matchRule` will actually
// match a section line. This is because `matchRule` does not verify
// whether a pattern is a valid pattern. A line like "[documentation]"
// would be considered a pattern without owners (which is supported).
func (p *parsing) matchRule() (string, []string, bool) {
	match := rulePattern.FindStringSubmatch(p.lineWithoutComments())
	if len(match) != 3 {
		return "", nil, false
	}
	filePattern := match[1]
	owners := strings.Fields(match[2])
	return filePattern, owners, true
}

var sectionPattern = lazyregexp.New(`^\s*\^?\s*\[([^\]]+)\]\s*(?:\[[0-9]+\])?\s*$`)

// matchSection tries to extract a section which looks like `[section name]`.
// A section can also be defined as `^[Section]`, meaning it is optional for approval.
// It can also be `[Section][2]`, meaning two approvals are required.
func (p *parsing) matchSection() bool {
	match := sectionPattern.FindStringSubmatch(p.lineWithoutComments())
	if len(match) != 2 {
		return false
	}
	p.section = match[1]
	return true
}

// isBlank returns true if the current line has no semantically relevant
// content. It can be blank while containing comments or whitespace.
func (p *parsing) isBlank() bool {
	return strings.TrimSpace(p.lineWithoutComments()) == ""
}

const (
	commentStart    = rune('#')
	escapeCharacter = rune('\\')
)

// lineWithoutComments returns the current line with the commented part
// stripped out.
func (p *parsing) lineWithoutComments() string {
	// A sensible default for index of the first byte where line-comment
	// starts is the line length. When the comment is removed by slicing
	// the string at the end, using the line-length as the index
	// of the first character dropped, yields the original string.
	commentStartIndex := len(p.line)
	var isEscaped bool
	for i, c := range p.line {
		// Unescaped # seen - this is where the comment starts.
		if c == commentStart && !isEscaped {
			commentStartIndex = i
			break
		}
		// Seeing escape character that is not being escaped itself (like \\)
		// means the following character is escaped.
		if c == escapeCharacter && !isEscaped {
			isEscaped = true
			continue
		}
		// Otherwise the next character is definitely not escaped.
		isEscaped = false
	}
	return p.line[:commentStartIndex]
}

func unescape(s string) string {
	var b strings.Builder
	var isEscaped bool
	for _, r := range s {
		if r == escapeCharacter && !isEscaped {
			isEscaped = true
			continue
		}
		b.WriteRune(r)
		isEscaped = false
	}
	return b.String()
}
