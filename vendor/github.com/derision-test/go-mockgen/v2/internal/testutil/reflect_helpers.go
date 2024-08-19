package testutil

import "reflect"

// CallInstance holds the arguments and results of a single mock function call.
type CallInstance interface {
	Args() []interface{}
	Results() []interface{}
}

// GetCallHistory extracts the history from the given mock function and returns the
// set of call instances.
func GetCallHistory(v interface{}) ([]CallInstance, bool) {
	value := reflect.ValueOf(v)
	if !value.IsValid() {
		return nil, false
	}

	// Get reflect value of method
	method := value.MethodByName("History")
	if !method.IsValid() {
		return nil, false
	}

	// Check method arity
	if method.Type().NumIn() != 0 || method.Type().NumOut() != 1 {
		return nil, false
	}

	// Invoke the function with no arguments and get the reflect.Value result
	history := method.Call(nil)[0]

	// Ensure the returned type is []interface{}
	if history.Kind() != reflect.Slice || !history.Type().Elem().Implements(reflect.TypeOf((*CallInstance)(nil)).Elem()) {
		return nil, false
	}

	calls := make([]CallInstance, 0, history.Len())
	for i := 0; i < history.Len(); i++ {
		calls = append(calls, history.Index(i).Interface().(CallInstance))
	}

	return calls, true
}

// GetCallHistoryWith extracts the history from the given mock function and returns the
// set of call instances that match the given function. If the given parameter is not of
// the required type, a false-valued flag is returned.
func GetCallHistoryWith(v interface{}, matcher func(v CallInstance) bool) (matching []CallInstance, _ bool) {
	history, ok := GetCallHistory(v)
	if !ok {
		return nil, false
	}

	for _, call := range history {
		if matcher(call) {
			matching = append(matching, call)
		}
	}

	return matching, true
}

// GetArgs returns the arguments from teh given mock function invocation. If the given
// parameter is not of the required type, a false-valued flag is returned.
func GetArgs(v interface{}) ([]interface{}, bool) {
	value := reflect.ValueOf(v)
	if !value.IsValid() {
		return nil, false
	}

	// Get reflect value of method
	method := value.MethodByName("Args")
	if !method.IsValid() {
		return nil, false
	}

	// Check method arity
	if method.Type().NumIn() != 0 || method.Type().NumOut() != 1 {
		return nil, false
	}

	// Invoke the function with no arguments and get the reflect.Value result
	args := method.Call(nil)[0]

	// Ensure the returned type is []interface{}
	if args.Kind() != reflect.Slice || args.Type().Elem().Kind() != reflect.Interface {
		return nil, false
	}

	// Return result unchanged
	return args.Interface().([]interface{}), true
}
