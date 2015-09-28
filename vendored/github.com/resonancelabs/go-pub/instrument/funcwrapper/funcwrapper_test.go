package funcwrapper

import (
	"fmt"
	"math"
	"testing"

	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument"
)

// (An arbitrary function)
func Divide(num, den float64) (float64, error) {
	if den == 0 {
		return math.Inf(1), fmt.Errorf("division by zero")
	}
	return num / den, nil
}

type S struct{}

func (s *S) SomeMethod(param string) {
	fmt.Printf("Param: %v\n", param)
}

// More sophisticated tests would be great but require access to the generated
// span/log information.
func TestWrapFuncInSpanCompilesAndIsNotTotallyBroken(t *testing.T) {
	// divWrapper must have the same type as the function it's wrapping
	// (`Divide` in this case).
	var divWrapper func(float64, float64) (float64, error)
	WrapFuncInSpan(
		Divide,
		&divWrapper,
		func(span instrument.ActiveSpan, inputs, outputs []interface{}) {
			_ = span.MergeTraceJoinIdsFromStack()
			span.SetEndUserId("whatever") // etc, etc
		})

	// We can now use divWrapper() just like Divide().
	q, err := divWrapper(7, 4)
	if q != 1.75 || err != nil {
		t.Error("Wrapper returned wrong result: ", q, err)
	}

	q, err = divWrapper(7, 0)
	if q != math.Inf(1) || err == nil {
		t.Error("Wrapper returned wrong result: ", q, err)
	}

	// Exercise the method (rather than function) case, too.
	s := S{}
	var someMethodWrapper func(string)
	WrapFuncInSpan(s.SomeMethod, &someMethodWrapper, nil)
	someMethodWrapper("it compiles, ship it!")
}

// More sophisticated tests would be great but require access to the generated
// span/log information.
func TestCallCompilesAndIsNotTotallyBroken(t *testing.T) {
	var q float64
	var err error
	Bind(Divide, 7, 4).DecorateAndCall(func(span instrument.ActiveSpan, i, o []interface{}) {
		fmt.Println("Inputs: ", i)
		fmt.Println("Outputs: ", o)
	}, &q, &err)
	if q != 1.75 || err != nil {
		t.Error("Wrapper returned wrong result: ", q, err)
	}

	// Check plain-old Call() too.
	Bind(Divide, 7, 4).Call(&q, &err)
	if q != 1.75 || err != nil {
		t.Error("Wrapper returned wrong result: ", q, err)
	}
}
