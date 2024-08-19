package mockassert

import (
	"reflect"

	"github.com/derision-test/go-mockgen/v2/internal/testutil"
	"github.com/stretchr/testify/assert"
)

// CallInstanceAsserter determines whether or not a set of argument values from a call
// of a mock function match the test constraints of a particular function call. See the
// assertions `CalledWith`, `NotCalledWith`, `CalledOnceWith`, `CalledNWith`, and
// `CalledAtNWith` for further usage.
type CallInstanceAsserter interface {
	// Assert determines if the given argument values matches the expected
	// function call.
	Assert(interface{}) bool
}

type CallInstanceAsserterFunc func(v interface{}) bool

func (f CallInstanceAsserterFunc) Assert(v interface{}) bool {
	return f(v)
}

type valueAsserter struct {
	expectedValues []interface{}
}

type skip struct{}

// Skip is a sentinel value which is skipped in a call instance asserter. This is useful
// when used to skip the leading "don't care" values such as leading context parameters.
var Skip = &skip{}

// Values returns a new call instance asserter that will match the arguments of each
// function call positionally with each of the expected values. The assertion behavior
// in each position can be tuned:
//
// Use the value `mockassert.Skip` to skip validation for values in that parameter
// position.
//
// Use a function with the type `func(v T) bool` (for any `T`) to override validation for
// values in that parameter position.
func Values(expectedValues ...interface{}) CallInstanceAsserter {
	return &valueAsserter{
		expectedValues: expectedValues,
	}
}

func (a *valueAsserter) Assert(v interface{}) bool {
	args, ok := testutil.GetArgs(v)
	if !ok {
		return false
	}

	if len(a.expectedValues) > len(args) {
		return false
	}

	for i, expectedValue := range a.expectedValues {
		if expectedValue == Skip {
			continue
		}

		// First check to see if it's a hook function we should invoke
		if ret, ok := callTesterFunc(expectedValue, args[i]); ok {
			if ret {
				continue
			}

			return false
		}

		// Fall back to value equality checks
		if assert.ObjectsAreEqual(expectedValue, args[i]) {
			continue
		}

		return false
	}

	return true
}

// callTesterFunc attempts to invoke the given value `v` of type func(T) bool
// with the given argument `arg` of type T.
//
// If the runtime types match these assumptions, then teh function is invoked
// and the result is returned along with a true-valued flag. If the runtime
// values break these assumptions, a false-valued flag is returned.
func callTesterFunc(v interface{}, arg interface{}) (result bool, ok bool) {
	value := reflect.ValueOf(v)
	if !value.IsValid() {
		return false, false
	}

	// Ensure value is a function (Type will panic otherwise)
	if value.Kind() != reflect.Func {
		return false, false
	}

	// Check function arity
	if value.Type().NumIn() != 1 || value.Type().NumOut() != 1 {
		return false, false
	}

	argValue := reflect.ValueOf(arg)

	// Ensure argument and parameter types match. (Call will panic otherwise)
	if value.Type().In(0).Kind() != argValue.Kind() {
		return false, false
	}

	// Invoke the function with a single argument and get the reflect.Value result
	resultValue := value.Call([]reflect.Value{argValue})[0]

	// Ensure the returned type is bool
	if resultValue.Kind() != reflect.Bool {
		return false, false
	}

	return resultValue.Interface().(bool), true
}
