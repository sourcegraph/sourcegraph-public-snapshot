package protocol

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
	proto "github.com/sourcegraph/sourcegraph/internal/searcher/v1"
)

type QueryNode interface {
	String() string
	ToProto() *proto.QueryNode
}

func FromJobNode(node query.Node) QueryNode {
	// In searcher, nil queries are represented as empty regex patterns
	if node == nil {
		return &PatternNode{
			Value:    "",
			IsRegExp: true,
		}
	}

	switch n := node.(type) {
	case query.Pattern:
		return &PatternNode{
			Value:     n.Value,
			IsNegated: n.Negated,
			IsRegExp:  n.Annotation.Labels.IsSet(query.Regexp),
		}
	case query.Operator:
		children := make([]QueryNode, 0, len(n.Operands))
		for _, operand := range n.Operands {
			children = append(children, FromJobNode(operand))
		}

		switch n.Kind {
		case query.And:
			return &AndNode{Children: children}
		case query.Or:
			return &OrNode{Children: children}
		default:
			panic(fmt.Sprintf("unexpected operator kind (%d)", n.Kind))
		}
	default:
		// Use a panic since this is used in a struct initializer, and it's really
		// verbose to pass through the error
		panic(fmt.Sprintf("unexpected query node type (%T)", n))
	}
}

type AndNode struct {
	Children []QueryNode
}

func (an *AndNode) String() string {
	return fmt.Sprintf("AND (%d children)", len(an.Children))
}

func (an *AndNode) ToProto() *proto.QueryNode {
	children := make([]*proto.QueryNode, 0, len(an.Children))
	for _, child := range an.Children {
		children = append(children, child.ToProto())
	}
	return &proto.QueryNode{
		Value: &proto.QueryNode_And{
			And: &proto.AndNode{Children: children},
		},
	}
}

type OrNode struct {
	Children []QueryNode
}

func (an *OrNode) String() string {
	return fmt.Sprintf("OR (%d children)", len(an.Children))
}

func (an *OrNode) ToProto() *proto.QueryNode {
	children := make([]*proto.QueryNode, 0, len(an.Children))
	for _, child := range an.Children {
		children = append(children, child.ToProto())
	}
	return &proto.QueryNode{
		Value: &proto.QueryNode_Or{
			Or: &proto.OrNode{Children: children},
		},
	}
}

type PatternNode struct {
	// Value is the search query. It is a regular expression if IsRegExp
	// is true, otherwise a fixed string. eg "route variable"
	Value string

	// IsNegated if true will invert the matching logic for regexp searches. IsNegated=true is
	// not supported for structural searches.
	IsNegated bool

	// IsRegExp if true will treat the Value as a regular expression.
	IsRegExp bool
}

func (rn *PatternNode) String() string {
	args := []string{}
	if rn.IsNegated {
		args = append(args, fmt.Sprintf("-%q", rn.Value))
	} else {
		args = append(args, fmt.Sprintf("%q", rn.Value))
	}

	if rn.IsRegExp {
		args = append(args, "re")
	}

	return fmt.Sprintf("PatternNode{%s}", strings.Join(args, ","))
}

func (rn *PatternNode) ToProto() *proto.QueryNode {
	return &proto.QueryNode{
		Value: &proto.QueryNode_Pattern{
			Pattern: &proto.PatternNode{
				Value:     rn.Value,
				IsNegated: rn.IsNegated,
				IsRegexp:  rn.IsRegExp,
			},
		},
	}
}
