package query

type Mapper interface {
	MapNodes(v Mapper, node []Node) []Node
	MapOperator(v Mapper, kind operatorKind, operands []Node) []Node
	MapParameter(v Mapper, field, value string, negated bool) Node
}

type BaseMapper struct{}

func (_ *BaseMapper) MapNodes(visitor Mapper, nodes []Node) []Node {
	mapped := []Node{}
	for _, node := range nodes {
		switch v := node.(type) {
		case Parameter:
			mapped = append(mapped, visitor.MapParameter(visitor, v.Field, v.Value, v.Negated))
		case Operator:
			mapped = append(mapped, visitor.MapOperator(visitor, v.Kind, v.Operands)...)
		}
	}
	return mapped
}

// Base mapper for Operators. Reduces operands if changed.
func (_ *BaseMapper) MapOperator(visitor Mapper, kind operatorKind, operands []Node) []Node {
	return newOperator(visitor.MapNodes(visitor, operands), kind)
}

// Base mapper for Parameters. It is the identity function.
func (_ *BaseMapper) MapParameter(visitor Mapper, field, value string, negated bool) Node {
	return Parameter{Field: field, Value: value, Negated: negated}
}

type ParameterMapper struct {
	callback func(field, value string, negated bool) Node
	BaseMapper
}

func (s *ParameterMapper) MapParameter(visitor Mapper, field, value string, negated bool) Node {
	return s.callback(field, value, negated)
}

// MapParameter calls callback on all parameter nodes. callback supplies the
// node's field, value, and whether the value is negated.
func MapParameter(nodes []Node, callback func(field, value string, negated bool) Node) []Node {
	visitor := &ParameterMapper{callback: callback}
	return visitor.MapNodes(visitor, nodes)
}
