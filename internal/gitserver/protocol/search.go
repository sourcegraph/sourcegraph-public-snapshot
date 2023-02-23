package protocol

import (
	"encoding/gob"
	"fmt"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Node interface {
	ToProto() *proto.QueryNode
	String() string
}

func NodeFromProto(p *proto.QueryNode) (Node, error) {
	switch v := p.GetValue().(type) {
	case *proto.QueryNode_AuthorMatches:
		return &AuthorMatches{Expr: v.AuthorMatches.Expr, IgnoreCase: v.AuthorMatches.IgnoreCase}, nil
	case *proto.QueryNode_CommitterMatches:
		return &CommitterMatches{Expr: v.CommitterMatches.Expr, IgnoreCase: v.CommitterMatches.IgnoreCase}, nil
	case *proto.QueryNode_CommitBefore:
		return &CommitBefore{Time: v.CommitBefore.GetTimestamp().AsTime()}, nil
	case *proto.QueryNode_CommitAfter:
		return &CommitAfter{Time: v.CommitAfter.GetTimestamp().AsTime()}, nil
	case *proto.QueryNode_MessageMatches:
		return &MessageMatches{Expr: v.MessageMatches.GetExpr(), IgnoreCase: v.MessageMatches.IgnoreCase}, nil
	case *proto.QueryNode_DiffMatches:
		return &DiffMatches{Expr: v.DiffMatches.GetExpr(), IgnoreCase: v.DiffMatches.IgnoreCase}, nil
	case *proto.QueryNode_DiffModifiesFile:
		return &DiffModifiesFile{Expr: v.DiffModifiesFile.GetExpr(), IgnoreCase: v.DiffModifiesFile.IgnoreCase}, nil
	case *proto.QueryNode_Boolean:
		return &Boolean{Value: v.Boolean.GetValue()}, nil
	case *proto.QueryNode_Operator:
		operands := make([]Node, 0, len(v.Operator.GetOperands()))
		for _, operand := range v.Operator.GetOperands() {
			node, err := NodeFromProto(operand)
			if err != nil {
				return nil, err
			}
			operands = append(operands, node)
		}
		var kind OperatorKind
		switch v.Operator.GetKind() {
		case proto.OperatorKind_OPERATOR_KIND_AND:
			kind = And
		case proto.OperatorKind_OPERATOR_KIND_OR:
			kind = Or
		case proto.OperatorKind_OPERATOR_KIND_NOT:
			kind = Not
		default:
			return nil, errors.Newf("unknown operator kind %s", v.Operator.GetKind().String())
		}
		return &Operator{Kind: kind, Operands: operands}, nil
	default:
		return nil, errors.Newf("unknown query node type %T", p.GetValue())
	}
}

// AuthorMatches is a predicate that matches if the author's name or email address
// matches the regex pattern.
type AuthorMatches struct {
	Expr       string
	IgnoreCase bool
}

func (a *AuthorMatches) String() string {
	return fmt.Sprintf("%T(%s)", a, a.Expr)
}

func (a *AuthorMatches) ToProto() *proto.QueryNode {
	return &proto.QueryNode{
		Value: &proto.QueryNode_AuthorMatches{
			AuthorMatches: &proto.AuthorMatchesNode{
				Expr:       a.Expr,
				IgnoreCase: a.IgnoreCase,
			},
		},
	}
}

// CommitterMatches is a predicate that matches if the author's name or email address
// matches the regex pattern.
type CommitterMatches struct {
	Expr       string
	IgnoreCase bool
}

func (c *CommitterMatches) String() string {
	return fmt.Sprintf("%T(%s)", c, c.Expr)
}

func (a *CommitterMatches) ToProto() *proto.QueryNode {
	return &proto.QueryNode{
		Value: &proto.QueryNode_CommitterMatches{
			CommitterMatches: &proto.CommitterMatchesNode{
				Expr:       a.Expr,
				IgnoreCase: a.IgnoreCase,
			},
		},
	}
}

// CommitBefore is a predicate that matches if the commit is before the given date
type CommitBefore struct {
	time.Time
}

func (c *CommitBefore) String() string {
	return fmt.Sprintf("%T(%s)", c, c.Time.String())
}

func (c *CommitBefore) ToProto() *proto.QueryNode {
	return &proto.QueryNode{
		Value: &proto.QueryNode_CommitBefore{
			CommitBefore: &proto.CommitBeforeNode{
				Timestamp: timestamppb.New(c.Time),
			},
		},
	}
}

// CommitAfter is a predicate that matches if the commit is after the given date
type CommitAfter struct {
	time.Time
}

func (c *CommitAfter) String() string {
	return fmt.Sprintf("%T(%s)", c, c.Time.String())
}

func (c *CommitAfter) ToProto() *proto.QueryNode {
	return &proto.QueryNode{
		Value: &proto.QueryNode_CommitAfter{
			CommitAfter: &proto.CommitAfterNode{
				Timestamp: timestamppb.New(c.Time),
			},
		},
	}
}

// MessageMatches is a predicate that matches if the commit message matches
// the provided regex pattern.
type MessageMatches struct {
	Expr       string
	IgnoreCase bool
}

func (m *MessageMatches) String() string {
	return fmt.Sprintf("%T(%s)", m, m.Expr)
}

func (m *MessageMatches) ToProto() *proto.QueryNode {
	return &proto.QueryNode{
		Value: &proto.QueryNode_MessageMatches{
			MessageMatches: &proto.MessageMatchesNode{
				Expr:       m.Expr,
				IgnoreCase: m.IgnoreCase,
			},
		},
	}
}

// DiffMatches is a a predicate that matches if any of the lines changed by
// the commit match the given regex pattern.
type DiffMatches struct {
	Expr       string
	IgnoreCase bool
}

func (d *DiffMatches) String() string {
	return fmt.Sprintf("%T(%s)", d, d.Expr)
}

func (m *DiffMatches) ToProto() *proto.QueryNode {
	return &proto.QueryNode{
		Value: &proto.QueryNode_DiffMatches{
			DiffMatches: &proto.DiffMatchesNode{
				Expr:       m.Expr,
				IgnoreCase: m.IgnoreCase,
			},
		},
	}
}

// DiffModifiesFile is a predicate that matches if the commit modifies any files
// that match the given regex pattern.
type DiffModifiesFile struct {
	Expr       string
	IgnoreCase bool
}

func (d *DiffModifiesFile) String() string {
	return fmt.Sprintf("%T(%s)", d, d.Expr)
}

func (m *DiffModifiesFile) ToProto() *proto.QueryNode {
	return &proto.QueryNode{
		Value: &proto.QueryNode_DiffModifiesFile{
			DiffModifiesFile: &proto.DiffModifiesFileNode{
				Expr:       m.Expr,
				IgnoreCase: m.IgnoreCase,
			},
		},
	}
}

// Boolean is a predicate that will either always match or never match
type Boolean struct {
	Value bool
}

func (c *Boolean) String() string {
	return fmt.Sprintf("%T(%t)", c, c.Value)
}

func (c *Boolean) ToProto() *proto.QueryNode {
	return &proto.QueryNode{
		Value: &proto.QueryNode_Boolean{
			Boolean: &proto.BooleanNode{
				Value: c.Value,
			},
		},
	}
}

type OperatorKind int

const (
	And OperatorKind = iota
	Or
	Not
)

func (o OperatorKind) ToProto() proto.OperatorKind {
	switch o {
	case And:
		return proto.OperatorKind_OPERATOR_KIND_AND
	case Or:
		return proto.OperatorKind_OPERATOR_KIND_OR
	case Not:
		return proto.OperatorKind_OPERATOR_KIND_NOT
	default:
		return proto.OperatorKind_OPERATOR_KIND_UNSPECIFIED
	}
}

type Operator struct {
	Kind     OperatorKind
	Operands []Node
}

func (o *Operator) String() string {
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

func (o *Operator) ToProto() *proto.QueryNode {
	operands := make([]*proto.QueryNode, 0, len(o.Operands))
	for _, operand := range o.Operands {
		operands = append(operands, operand.ToProto())
	}
	return &proto.QueryNode{
		Value: &proto.QueryNode_Operator{
			Operator: &proto.OperatorNode{
				Kind:     o.Kind.ToProto(),
				Operands: operands,
			},
		},
	}
}

// newOperator is a convenience function for internal construction of operators.
// It does no simplification of its arguments, so generally should not be used
// by consumers directly. Prefer NewAnd, NewOr, and NewNot.
func newOperator(kind OperatorKind, operands ...Node) *Operator {
	return &Operator{
		Kind:     kind,
		Operands: operands,
	}
}

// NewAnd creates a new And node from the given operands
// Optimizations/simplifications:
// - And() => Boolean(true)
// - And(x) => x
// - And(x, And(y, z)) => And(x, y, z)
func NewAnd(operands ...Node) Node {
	// An empty And operator will always match a commit
	if len(operands) == 0 {
		return &Boolean{true}
	}

	// An And operator with a single operand can be unwrapped
	if len(operands) == 1 {
		return operands[0]
	}

	// Flatten any nested And operands since And is associative
	// P ∧ (Q ∧ R) <=> (P ∧ Q) ∧ R
	flattened := make([]Node, 0, len(operands))
	for _, operand := range operands {
		if nestedOperator, ok := operand.(*Operator); ok && nestedOperator.Kind == And {
			flattened = append(flattened, nestedOperator.Operands...)
		} else {
			flattened = append(flattened, operand)
		}
	}

	return newOperator(And, flattened...)
}

// NewOr creates a new Or node from the given operands.
// Optimizations/simplifications:
// - Or() => Boolean(false)
// - Or(x) => x
// - Or(x, Or(y, z)) => Or(x, y, z)
func NewOr(operands ...Node) Node {
	// An empty Or operator will never match a commit
	if len(operands) == 0 {
		return &Boolean{false}
	}

	// An Or operator with a single operand can be unwrapped
	if len(operands) == 1 {
		return operands[0]
	}

	// Flatten any nested Or operands
	flattened := make([]Node, 0, len(operands))
	for _, operand := range operands {
		if nestedOperator, ok := operand.(*Operator); ok && nestedOperator.Kind == Or {
			flattened = append(flattened, nestedOperator.Operands...)
		} else {
			flattened = append(flattened, operand)
		}
	}

	return newOperator(Or, flattened...)
}

// NewNot creates a new negated node from the given operand
// Optimizations/simplifications:
// - Not(Not(x)) => x
func NewNot(operand Node) Node {
	// If an operator, push the negation down to the atom nodes recursively
	if operator, ok := operand.(*Operator); ok && operator.Kind == Not {
		return operator.Operands[0]
	}

	// If an atom node, just negate it
	return newOperator(Not, operand)
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
		gob.Register(&Boolean{})
		gob.Register(&Operator{})
	})
}
