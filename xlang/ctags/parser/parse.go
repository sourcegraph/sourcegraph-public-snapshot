package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Tag is a Go representation of a line in a tags file. Read man 1 ctags for a
// description of these fields.
type Tag struct {
	// File is the relative file path.
	File string

	// DefLinePrefix is the ex command to find the line.
	DefLinePrefix string

	// Name is the name of the identifier.
	Name string

	// Linenumber is the number of newlines between the beginning of the file
	// and the identifier.
	LineNumber int

	// Kind can be class, func, etc. and varies by language.
	Kind string

	// Signature is the static type of the identifier
	Signature string

	// Language is the language of the identifier
	Language string
}

type TagsParser struct {
	// tags is the Go representation of a ctags file.
	tags []Tag

	// curFile contains the file being parsed.
	curFile string
}

func Parse(r *bufio.Reader) ([]Tag, error) {
	p := TagsParser{}
	p.curFile = ""

	line, err := r.ReadString('\n')
	for ; err == nil; line, err = r.ReadString('\n') {
		if err := p.parseLine(strings.TrimRight(line, "\r\n")); err != nil {
			return nil, err
		}
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	return p.tags, nil
}

func (p *TagsParser) parseLine(line string) error {
	if len(strings.TrimSpace(line)) == 0 || strings.HasPrefix(line, "!") {
		return nil
	}

	t1 := strings.Index(line, "\t")
	if t1 == -1 {
		return fmt.Errorf("expected tab-delimited line with at least 4 fields, but got %q", line)
	}
	name := line[0:t1]

	offset := strings.Index(line[t1+1:], "\t")
	if offset == -1 {
		return fmt.Errorf("expected tab-delimited line with at least 4 fields, but got %q", line)
	}
	t2 := t1 + 1 + offset
	file := line[t1+1 : t2]

	offset = strings.LastIndex(line[t2+1:], `;"`)
	if offset == -1 {
		return fmt.Errorf(`expected find command to terminate with ';"', but got %q`, line)
	}
	t3 := offset + 2 + t2 + 1
	if len(line) <= t3 || line[t3] != '\t' {
		return fmt.Errorf(`expected tab immediately following ';"', but got %q, line: was %q`, line[t3:t3+1], line)
	}
	findCmd := line[t2+1 : t3]

	extensionFields := strings.Split(line[t3+1:], "\t")
	fields := make(map[string]string)
	for _, field := range extensionFields {
		s := strings.Index(field, ":")
		key, val := field[0:s], field[s+1:]
		fields[key] = val
	}
	lineno, err := strconv.Atoi(fields["line"])
	if err != nil {
		return fmt.Errorf("could not parse line number, line was %q", line)
	}

	p.tags = append(p.tags, Tag{
		Name:          name,
		File:          file,
		DefLinePrefix: findCmdToDefLinePrefix(findCmd),
		Kind:          fields["kind"],
		Signature:     fields["signature"],
		Language:      fields["language"],
		LineNumber:    lineno,
	})
	return nil
}

func findCmdToDefLinePrefix(findCmd string) string {
	def := strings.TrimSuffix(strings.TrimPrefix(findCmd, `/^`), `/;"`)
	if strings.HasSuffix(def, "$") {
		def = strings.TrimSuffix(def, "$")
	}
	return def
}
