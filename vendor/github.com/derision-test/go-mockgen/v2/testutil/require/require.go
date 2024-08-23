package mockrequire

import (
	mockassert "github.com/derision-test/go-mockgen/v2/testutil/assert"
	"github.com/stretchr/testify/require"
)

// Called asserts that the mock function object was called at least once.
func Called(t require.TestingT, mockFn interface{}, msgAndArgs ...interface{}) {
	if !mockassert.Called(t, mockFn, msgAndArgs...) {
		t.FailNow()
	}
}

// NotCalled asserts that the mock function object was not called.
func NotCalled(t require.TestingT, mockFn interface{}, msgAndArgs ...interface{}) {
	if !mockassert.NotCalled(t, mockFn, msgAndArgs...) {
		t.FailNow()
	}
}

// CalledOnce asserts that the mock function object was called exactly once.
func CalledOnce(t require.TestingT, mockFn interface{}, msgAndArgs ...interface{}) {
	if !mockassert.CalledOnce(t, mockFn, msgAndArgs...) {
		t.FailNow()
	}
}

// CalledN asserts that the mock function object was called exactly n times.
func CalledN(t require.TestingT, mockFn interface{}, n int, msgAndArgs ...interface{}) {
	if !mockassert.CalledN(t, mockFn, n, msgAndArgs...) {
		t.FailNow()
	}
}

// CalledWith asserts that the mock function object was called at least once with a set of
// arguments matching the given mockassertion function.
func CalledWith(t require.TestingT, mockFn interface{}, asserter CallInstanceAsserter, msgAndArgs ...interface{}) {
	if !mockassert.CalledWith(t, mockFn, asserter, msgAndArgs...) {
		t.FailNow()
	}
}

// NotCalledWith asserts that the mock function object was not called with a set of arguments
// matching the given mockassertion function.
func NotCalledWith(t require.TestingT, mockFn interface{}, asserter CallInstanceAsserter, msgAndArgs ...interface{}) {
	if !mockassert.NotCalledWith(t, mockFn, asserter, msgAndArgs...) {
		t.FailNow()
	}
}

// CalledOnceWith asserts that the mock function object was called exactly once with a set of
// arguments matching the given mockassertion function.
func CalledOnceWith(t require.TestingT, mockFn interface{}, asserter CallInstanceAsserter, msgAndArgs ...interface{}) {
	if !mockassert.CalledOnceWith(t, mockFn, asserter, msgAndArgs...) {
		t.FailNow()
	}
}

// CalledNWith asserts that the mock function object was called exactly n times with a set of
// arguments matching the given mockassertion function.
func CalledNWith(t require.TestingT, mockFn interface{}, n int, asserter CallInstanceAsserter, msgAndArgs ...interface{}) {
	if !mockassert.CalledNWith(t, mockFn, n, asserter, msgAndArgs...) {
		t.FailNow()
	}
}

// CalledAtNWith asserts that the mock function objects nth call was with a set of
// arguments matching the given call instance asserter.
func CalledAtNWith(t require.TestingT, mockFn interface{}, n int, asserter CallInstanceAsserter, msgAndArgs ...interface{}) {
	if !mockassert.CalledAtNWith(t, mockFn, n, asserter, msgAndArgs...) {
		t.FailNow()
	}
}
