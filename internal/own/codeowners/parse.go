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
		if p.blank() {
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

type parsing struct {
	line string
}

// nextRule advances parsing to focus on the next line.
// Conveniently returns the same object for chaining with `notBlank()`.
func (p *parsing) nextLine(line string) {
	p.line = line
}

var (
	rulePattern = regexp.MustCompile(`^\s*(\S+)((?:\s+\S+)*)\s*$`)
)

// matchRule tries to extract a codeowners rule from current line
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

func (p *parsing) blank() bool {
	return strings.TrimSpace(p.lineWithoutComments()) == ""
}

func (p *parsing) lineWithoutComments() string {
	return p.line[:p.commentStartIndex()]
}

const (
	commentStart    = rune('#')
	escapeCharacter = rune('\\')
)

func (p *parsing) commentStartIndex() int {
	var esc bool
	for i, c := range p.line {
		if c == commentStart && !esc {
			return i
		}
		if c == escapeCharacter && !esc {
			esc = true
			continue
		}
		if esc {
			esc = false
		}
	}
	return len(p.line)
}
