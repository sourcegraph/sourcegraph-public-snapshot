package codeowners

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/proto"
)

func Parse(in string) (*codeownerspb.File, error) {
	return Read(strings.NewReader(in))
}

func Read(in io.Reader) (*codeownerspb.File, error) {
	scanner := bufio.NewScanner(in)
	var rs []*codeownerspb.Rule
	p := new(parsing)
	for scanner.Scan() {
		p.nextLine(scanner.Text())
		if p.isBlank() {
			continue
		}
		pattern, owners, ok := p.matchRule()
		if !ok {
			return nil, fmt.Errorf("failed to match rule: %s", p.line)
		}
		r := codeownerspb.Rule{Pattern: pattern}
		for _, ownerText := range owners {
			var o codeownerspb.Owner
			if strings.HasPrefix(ownerText, "@") {
				o.Handle = strings.TrimPrefix(ownerText, "@")
			} else {
				// Note: we assume owner text is an email if it does not
				// start with an `@` which would make it a handle.
				o.Email = ownerText
			}
			r.Owner = append(r.Owner, &o)
		}
		rs = append(rs, &r)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &codeownerspb.File{Rule: rs}, nil
}

// parsing implements matching and parsing primitives for CODEOWNERS files
// as well as keeps track of internal state as a file is being parsed.
type parsing struct {
	// line is the current line being parsed. CODEOWNERS files are built
	// in such a way that for syntactic purposes, every line can be considered
	// in isolation.
	line string
}

// nextLine advances parsing to focus on the next line.
// Conveniently returns the same object for chaining with `notBlank()`.
func (p *parsing) nextLine(line string) {
	p.line = line
}

var (
	// rulePattern is expected to match a rule line like:
	// `cmd/**/docs/index.md @readme-owners owner@example.com`.
	//  ^^^^^^^^^^^^^^^^^^^^ ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
	// The first capturing   The second capturing group
	// group extracts        extracts all the owners
	// the file pattern.     separated by whitespace.
	rulePattern = regexp.MustCompile(`^\s*(\S+)((?:\s+\S+)*)\s*$`)
)

// matchRule tries to extract a codeowners rule from the current line
// and return the file pattern and one or more owners.
// Match is indicated by the third return value being true.
func (p *parsing) matchRule() (string, []string, bool) {
	match := rulePattern.FindStringSubmatch(p.lineWithoutComments())
	if len(match) != 3 {
		return "", nil, false
	}
	filePattern := match[1]
	owners := strings.Fields(match[2])
	return filePattern, owners, true
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
	// starts is the line lenght. When the comment is removed by slcing
	// the string at the end, using the line-length as the index
	// of the first character dropped, yields the original string.
	commentStartIndex := len(p.line)
	var esc bool // whether current character is escaped.
	for i, c := range p.line {
		// Unespcaped # seen - this is where the comment starts.
		if c == commentStart && !esc {
			commentStartIndex = i
			break
		}
		// Seeing escape character that is not being escaped itself (like \\)
		// means the following character is escaped.
		if c == escapeCharacter && !esc {
			esc = true
			continue
		}
		// Otherwise the next character is definitely not escaped.
		esc = false

	}
	return p.line[:commentStartIndex]
}
