package protocol

import (
	"encoding/gob"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

type SearchQuery interface {
	String() string
}

// AuthorMatches is a predicate that matches if the author's name or email address
// matches the regex pattern.
type AuthorMatches struct {
	Regexp
}

func (a AuthorMatches) String() string {
	return fmt.Sprintf("%T(%s)", a, a.Regexp.String())
}

// CommitterMatches is a predicate that matches if the author's name or email address
// matches the regex pattern.
type CommitterMatches struct {
	Regexp
}

func (c CommitterMatches) String() string {
	return fmt.Sprintf("%T(%s)", c, c.Regexp.String())
}

// CommitBefore is a predicate that matches if the commit is before the given date
type CommitBefore struct {
	time.Time
}

func (c CommitBefore) String() string {
	return fmt.Sprintf("%T(%s)", c, c.Time.String())
}

// CommitAfter is a predicate that matches if the commit is after the given date
type CommitAfter struct {
	time.Time
}

func (c CommitAfter) String() string {
	return fmt.Sprintf("%T(%s)", c, c.Time.String())
}

// MessageMatches is a predicate that matches if the commit message matches
// the provided regex pattern.
type MessageMatches struct {
	Regexp
}

func (m MessageMatches) String() string {
	return fmt.Sprintf("%T(%s)", m, m.Regexp.String())
}

// DiffMatches is a a predicate that matches if any of the lines changed by
// the commit match the given regex pattern.
type DiffMatches struct {
	Regexp
}

func (d DiffMatches) String() string {
	return fmt.Sprintf("%T(%s)", d, d.Regexp.String())
}

// DiffModifiesFile is a predicate that matches if the commit modifies any files
// that match the given regex pattern.
type DiffModifiesFile struct {
	Regexp
}

func (d DiffModifiesFile) String() string {
	return fmt.Sprintf("%T(%s)", d, d.Regexp.String())
}

// And is a predicate that matches if all of its children predicates match
type And struct {
	Children []SearchQuery
}

func (a And) String() string {
	cs := make([]string, 0, len(a.Children))
	for _, child := range a.Children {
		cs = append(cs, child.String())
	}
	return "(" + strings.Join(cs, " AND ") + ")"
}

// Or is a predicate that matches if any of its children predicates match
type Or struct {
	Children []SearchQuery
}

func (o Or) String() string {
	cs := make([]string, 0, len(o.Children))
	for _, child := range o.Children {
		cs = append(cs, child.String())
	}
	return "(" + strings.Join(cs, " OR ") + ")"
}

// Not is a predicate that matches if its child predicate does not match
type Not struct {
	Child SearchQuery
}

func (n Not) String() string {
	return "NOT " + n.Child.String()
}

// Regexp is a thin wrapper around the stdlib Regexp type that enables gob encoding
type Regexp struct {
	*regexp.Regexp
}

func (r Regexp) GobEncode() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r *Regexp) GobDecode(data []byte) (err error) {
	r.Regexp, err = regexp.Compile(string(data))
	return err
}

var registerOnce sync.Once

func RegisterGob() {
	registerOnce.Do(func() {
		gob.Register(&AuthorMatches{})
		gob.Register(&CommitterMatches{})
		gob.Register(&CommitBefore{})
		gob.Register(&CommitAfter{})
		gob.Register(&MessageMatches{})
		gob.Register(&DiffMatches{})
		gob.Register(&DiffModifiesFile{})
		gob.Register(&And{})
		gob.Register(&Or{})
		gob.Register(&Not{})
	})
}
