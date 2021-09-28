package protocol

import (
	"encoding/gob"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Node interface {
	String() string
}

// AuthorMatches is a predicate that matches if the author's name or email address
// matches the regex pattern.
type AuthorMatches struct {
	Expr       string
	IgnoreCase bool
}

func (a AuthorMatches) String() string {
	return fmt.Sprintf("%T(%s)", a, a.Expr)
}

// CommitterMatches is a predicate that matches if the author's name or email address
// matches the regex pattern.
type CommitterMatches struct {
	Expr       string
	IgnoreCase bool
}

func (c CommitterMatches) String() string {
	return fmt.Sprintf("%T(%s)", c, c.Expr)
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
	Expr       string
	IgnoreCase bool
}

func (m MessageMatches) String() string {
	return fmt.Sprintf("%T(%s)", m, m.Expr)
}

// DiffMatches is a a predicate that matches if any of the lines changed by
// the commit match the given regex pattern.
type DiffMatches struct {
	Expr       string
	IgnoreCase bool
}

func (d DiffMatches) String() string {
	return fmt.Sprintf("%T(%s)", d, d.Expr)
}

// DiffModifiesFile is a predicate that matches if the commit modifies any files
// that match the given regex pattern.
type DiffModifiesFile struct {
	Expr       string
	IgnoreCase bool
}

func (d DiffModifiesFile) String() string {
	return fmt.Sprintf("%T(%s)", d, d.Expr)
}

type OperatorKind int

const (
	And OperatorKind = iota
	Or
	Not
)

type Operator struct {
	Kind     OperatorKind
	Operands []Node
}

func (o Operator) String() string {
	var sep, prefix string
	switch o.Kind {
	case And:
		sep = " AND "
	case Or:
		sep = " OR "
	case Not:
		sep = " AND NOT "
		prefix = "NOT "
	}

	cs := make([]string, 0, len(o.Operands))
	for _, operand := range o.Operands {
		cs = append(cs, operand.String())
	}
	return "(" + prefix + strings.Join(cs, sep) + ")"
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
		gob.Register(&Operator{})
	})
}
