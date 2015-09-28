package funcwrapper

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument"
)

type DecoratorFunc func(span instrument.ActiveSpan, inputs, outputs []interface{})

// See Closure.Call().
type Closure struct {
	// boundFunc.Type().Kind()==reflect.Func
	boundFunc reflect.Value

	// The bound parameters for the invocation of `boundFunc`
	boundParamVals   []reflect.Value
	boundParamIfaces []interface{}
}

// See Closure.Call().
//
// `wrappedFunc` should be a function (or method bound to a particular
// instance) to wrap.
//
// There must be an entry in the variadic `boundParams` for each input
// parameter to `wrappedFunc`. Each parameter must be convertible to the
// corresponding input parameter type for `wrappedFunc` (just as it would be in
// a normal golang function invocation).
//
// Note that the wrapped function is *not* invoked until the programmer invokes
// Closure.Call() on the returned *Closure.
func Bind(wrappedFunc interface{}, boundParams ...interface{}) *Closure {
	fVal := reflect.ValueOf(wrappedFunc)
	if len(boundParams) != fVal.Type().NumIn() {
		panic(fmt.Errorf("Wrong number of bound parameters"))
	}

	bpVals := make([]reflect.Value, len(boundParams))
	for i, bp := range boundParams {
		bpVals[i] = reflect.ValueOf(bp).Convert(fVal.Type().In(i))
	}
	return &Closure{fVal, bpVals, boundParams}
}

// Bind() and Closure.Call() allow the programmer to wrap a Span around an
// arbitrary function call while preserving type information.
//
// Bind() associates a function with its [input] parameters and returns a new
// *Closure. Closure.Call() invokes the function with its bound parameters and
// returns the zero or more output values via `retValPtrs`.
//
// `decorator` may be nil. If specified, it is given access to all input and
// output parameters for this particular Call() invocation. If nil,
// StackSpanDecorator is used as a default.
//
// There must be an entry in the variadic `retValPtrs` for each return value of
// the function passed to Bind(). Each entry should be a pointer to the type of
// the respective return value.
//
// The implementation will panic if types passed in for parameter or return
// value bindings are not compatible with the wrapped function.
//
// For example:
//
//    // (An arbitrary function)
//    func Divide(num, den float64) (float64, error) {
//        if den == 0 {
//            return math.Inf(1), fmt.Errorf("division by zero")
//        }
//        return num / den, nil
//    }
//
//    ...
//
//    var q float64
//    var err error
//    funcwrapper.Bind(Divide, 7, 4).Call(
//        &q, &err)  // note that these are pointers!
//    fmt.Printf("%v, %v", q, err)  // prints "1.75, nil"
//
func (c *Closure) Call(retValPtrs ...interface{}) {
	c.DecorateAndCall(nil, retValPtrs...)
}
func (c *Closure) DecorateAndCall(decorator DecoratorFunc, retValPtrs ...interface{}) {
	wrappedPackageName, wrappedFuncName := getPackageAndFunctionNames(c.boundFunc)
	wrapperSpan := instrument.StartSpan().SetOperation(
		"funcwrapper/" + wrappedPackageName + "/" + wrappedFuncName)
	defer wrapperSpan.Finish()

	wrapperSpan.Log(instrument.Println("input values").Payload(c.boundParamIfaces))
	rvals := c.boundFunc.Call(c.boundParamVals)

	if len(rvals) != len(retValPtrs) {
		panic(fmt.Errorf("Expected %d return values, got %d", len(retValPtrs), len(rvals)))
	}
	outputIfaces := make([]interface{}, len(rvals))
	for i, actualRetVal := range rvals {
		destElem := reflect.ValueOf(retValPtrs[i]).Elem()
		destElem.Set(actualRetVal.Convert(destElem.Type()))
		outputIfaces[i] = destElem.Interface()
	}
	wrapperSpan.Log(instrument.Println("output values").Payload(outputIfaces))
	if decorator == nil {
		decorator = StackSpanDecorator
	}
	decorator(wrapperSpan, c.boundParamIfaces, outputIfaces)
}

// Return (via an output param) a new function that wraps an arbitrary
// function, `toWrap`, in a span that:
//  - has the same name as `toWrap`
//  - logs all param and return values as payload structs
//  - if `decorateSpan` is not nil, allows the caller to invoke arbitrary
//    methods on the wrapper span (to set TraceJoinIds, etc).
//
// The wrapped function is stored in the `*wrapperPtr` parameter. `wrapperPtr`
// must be a pointer to a function with the same signature as `toWrap`. Go
// needs this to make the wrapper function behave like the wrapped function
// from a type-checking standpoint.
//
// `decorator` may be nil. If specified, it is given access to all input and
// output parameters for this particular Call() invocation. If nil,
// StackSpanDecorator is used as a default.
//
// For example:
//
//    // (An arbitrary function)
//    func Divide(num, den float64) (float64, error) {
//        if den == 0 {
//            return math.Inf(1), fmt.Errorf("division by zero")
//        }
//        return num / den, nil
//    }
//
//    ...
//
//    // divWrapper must have the same type as the function it's wrapping
//    // (`Divide` in this case).
//    var divWrapper func(float64, float64) (float64, error)
//    funcwrapper.WrapFuncInSpan(
//        Divide,
//        &divWrapper,
//        func(span instrument.ActiveSpan, inputs, outputs []reflect.Value) {
//        	  _ = span.MergeTraceJoinIdsFromStack()
//        	  span.SetEndUserId(User.Username())  // etc, etc
//        })
//
//    // We can now use divWrapper() just like Divide().
//    q, err := divWrapper(7, 4)
//    fmt.Printf("%v, %v", q, err)  // prints "1.75, nil"
//
func WrapFuncInSpan(
	toWrap interface{},
	wrapperPtr interface{},
	decorator DecoratorFunc) {
	// (The implementation borrows from the reflect.MakeFunc godoc example.)

	// Obtain the function value itself (likely nil) as a reflect.Value
	// so that we can query its type and then set the value.
	wrapperFn := reflect.ValueOf(wrapperPtr).Elem()

	toWrapVal := reflect.ValueOf(toWrap)
	wrappedPackageName, wrappedFuncName := getPackageAndFunctionNames(toWrapVal)

	// Make a function of the right type.
	v := reflect.MakeFunc(wrapperFn.Type(), func(inVals []reflect.Value) []reflect.Value {
		wrapperSpan := instrument.StartSpan().SetOperation(
			"funcwrapper/" + wrappedPackageName + "/" + wrappedFuncName)

		// Interpose on the inputs.
		inIfaces := make([]interface{}, len(inVals))
		for i, v := range inVals {
			inIfaces[i] = v.Interface()
		}
		wrapperSpan.Log(instrument.Println("param values").Payload(inIfaces))

		// Actually call 'toWrap'.
		outVals := toWrapVal.Call(inVals)

		// Interpose on the outputs.
		outIfaces := make([]interface{}, len(outVals))
		for i, v := range outVals {
			outIfaces[i] = v.Interface()
		}
		wrapperSpan.Log(instrument.Println("return values").Payload(inIfaces))

		// Now that we have both inputs and outputs, call the user span
		// decoration code (if any).
		if decorator == nil {
			decorator = StackSpanDecorator
		}
		decorator(wrapperSpan, inIfaces, outIfaces)
		wrapperSpan.Finish()

		// Return the wrapped outputs as part of the interposition contract.
		return outVals
	})

	// Assign it to the value wrapperFn represents.
	wrapperFn.Set(v)
}

// A simple common-case helper that can be passed as the `decorateSpan`
// parameter of `WrapFuncInSpan()`.
func StackSpanDecorator(span instrument.ActiveSpan, in, out []interface{}) {
	_ = span.MergeTraceJoinIdsFromStack()
}

func getRawFunctionName(fVal reflect.Value) string {
	var rval string
	defer func() {
		if r := recover(); r != nil {
			rval = "?.?"
		}
	}()
	// May panic; see the defer above.
	rval = runtime.FuncForPC(fVal.Pointer()).Name()
	return rval
}

func getPackageAndFunctionNames(fVal reflect.Value) (string, string) {
	rawName := getRawFunctionName(fVal)
	idx := strings.LastIndex(rawName, ".")
	if idx < 0 || idx >= len(rawName) {
		return "?", rawName
	}
	return rawName[:idx], rawName[idx+1:]
}
