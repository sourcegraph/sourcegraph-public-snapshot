package redigomock

import "reflect"

// FuzzyMatcher is an interface that exports one function. It can be
// passed to the Command as an argument. When the command is evaluated against
// data provided in mock connection Do call, FuzzyMatcher will call Match on the
// argument and return true if the argument fulfills constraints set in concrete
// implementation
type FuzzyMatcher interface {

	// Match takes an argument passed to mock connection Do call and checks if
	// it fulfills constraints set in concrete implementation of this interface
	Match(interface{}) bool
}

// NewAnyInt returns a FuzzyMatcher instance matching any integer passed as an
// argument
func NewAnyInt() FuzzyMatcher {
	return anyInt{}
}

// NewAnyDouble returns a FuzzyMatcher instance matching any double passed as
// an argument
func NewAnyDouble() FuzzyMatcher {
	return anyDouble{}
}

// NewAnyData returns a FuzzyMatcher instance matching every data type passed
// as an argument (returns true by default)
func NewAnyData() FuzzyMatcher {
	return anyData{}
}

type anyInt struct{}

func (matcher anyInt) Match(input interface{}) bool {
	switch input.(type) {
	case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
		return true
	default:
		return false
	}
}

type anyDouble struct{}

func (matcher anyDouble) Match(input interface{}) bool {
	switch input.(type) {
	case float32, float64:
		return true
	default:
		return false
	}
}

type anyData struct{}

func (matcher anyData) Match(input interface{}) bool {
	return true
}

func implementsFuzzy(input interface{}) bool {
	inputType := reflect.TypeOf(input)
	if inputType == nil {
		return false
	}
	return inputType.Implements(reflect.TypeOf((*FuzzyMatcher)(nil)).Elem())
}
