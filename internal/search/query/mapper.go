package query

// The Mapper interface allows to replace nodes for each respective part of the
// query grammar. It is a visitor that will replace the visited node by the
// returned value.
type Mapper interface {
	MapNodes(v Mapper, node []Node) []Node
	MapOperator(v Mapper, kind operatorKind, operands []Node) []Node
	MapParameter(v Mapper, field, value string, negated bool) Node
	MapPattern(v Mapper, value string, negated, quoted bool) Node
}

// The BaseMapper is a mapper that recursively visits each node in a query and
// maps it to itself. A BaseMapper's methods may be overrided by embedding it a
// custom mapper's definitoin. See ParameterMapper for an example.
type BaseMapper struct{}

func (*BaseMapper) MapNodes(visitor Mapper, nodes []Node) []Node {
	mapped := []Node{}
	for _, node := range nodes {
		switch v := node.(type) {
		case Pattern:
			mapped = append(mapped, visitor.MapPattern(visitor, v.Value, v.Negated, v.Quoted))
		case Parameter:
			mapped = append(mapped, visitor.MapParameter(visitor, v.Field, v.Value, v.Negated))
		case Operator:
			mapped = append(mapped, visitor.MapOperator(visitor, v.Kind, v.Operands)...)
		}
	}
	return mapped
}

// Base mapper for Operators. Reduces operands if changed.
func (*BaseMapper) MapOperator(visitor Mapper, kind operatorKind, operands []Node) []Node {
	return newOperator(visitor.MapNodes(visitor, operands), kind)
}

// Base mapper for Parameters. It is the identity function.
func (*BaseMapper) MapParameter(visitor Mapper, field, value string, negated bool) Node {
	return Parameter{Field: field, Value: value, Negated: negated}
}

// Base mapper for Patterns. It is the identity function.
func (*BaseMapper) MapPattern(visitor Mapper, value string, negated, quoted bool) Node {
	return Pattern{Value: value, Negated: negated, Quoted: quoted}
}

// ParameterMapper is a helper mapper that only maps parameters in a query. It
// takes as state a callback that will call and map each visited parameter by
// the return value.
type ParameterMapper struct {
	callback func(field, value string, negated bool) Node
	BaseMapper
}

// MapParameter implements ParameterMapper by overriding the BaseMapper's value
// to substitute a node as determined by the callback.
func (s *ParameterMapper) MapParameter(visitor Mapper, field, value string, negated bool) Node {
	return s.callback(field, value, negated)
}

// MapParameter calls callback on all parameter nodes, substituting them for
// callback's return value. callback supplies the node's field, value, and
// whether the value is negated.
func MapParameter(nodes []Node, callback func(field, value string, negated bool) Node) []Node {
	visitor := &ParameterMapper{callback: callback}
	return visitor.MapNodes(visitor, nodes)
}
