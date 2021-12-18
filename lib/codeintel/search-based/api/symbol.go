package api

type Symbol struct {
	Value string
}

func NewSymbol(value string) *Symbol {
	return &Symbol{value}
}

func (s *Symbol) String() string {
	return s.Value
}

func (s *Symbol) IsEqual(other *Symbol) bool {
	return s.Value == other.Value
}
