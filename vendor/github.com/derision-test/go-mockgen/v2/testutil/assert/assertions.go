package mockassert

import (
	"fmt"

	"github.com/derision-test/go-mockgen/v2/internal/testutil"
	"github.com/stretchr/testify/assert"
)

// Called asserts that the mock function object was called at least once.
func Called(t assert.TestingT, mockFn interface{}, msgAndArgs ...interface{}) bool {
	callCount, ok := callCount(t, mockFn, msgAndArgs...)
	if !ok {
		return false
	}
	if callCount == 0 {
		return assert.Fail(t, fmt.Sprintf("Expected %T to be called at least once", mockFn), msgAndArgs...)
	}

	return true
}

// NotCalled asserts that the mock function object was not called.
func NotCalled(t assert.TestingT, mockFn interface{}, msgAndArgs ...interface{}) bool {
	callCount, ok := callCount(t, mockFn, msgAndArgs...)
	if !ok {
		return false
	}
	if callCount != 0 {
		return assert.Fail(t, fmt.Sprintf("Did not expect %T to be called", mockFn), msgAndArgs...)
	}

	return true
}

// CalledOnce asserts that the mock function object was called exactly once.
func CalledOnce(t assert.TestingT, mockFn interface{}, msgAndArgs ...interface{}) bool {
	return CalledN(t, mockFn, 1, msgAndArgs...)
}

// CalledN asserts that the mock function object was called exactly n times.
func CalledN(t assert.TestingT, mockFn interface{}, n int, msgAndArgs ...interface{}) bool {
	callCount, ok := callCount(t, mockFn, msgAndArgs...)
	if !ok {
		return false
	}
	if callCount != n {
		return assert.Fail(t, fmt.Sprintf("Expected %T to be called exactly %d times, called %d times", mockFn, n, callCount), msgAndArgs...)
	}

	return true
}

// CalledWith asserts that the mock function object was called at least once with a set of
// arguments matching the given call instance asserter.
func CalledWith(t assert.TestingT, mockFn interface{}, asserter CallInstanceAsserter, msgAndArgs ...interface{}) bool {
	matchingCallCount, ok := callCountWith(t, mockFn, asserter, msgAndArgs...)
	if !ok {
		return false
	}
	if matchingCallCount == 0 {
		return assert.Fail(t, fmt.Sprintf("Expected %T to be called with given arguments at least once", mockFn), msgAndArgs...)
	}
	return true
}

// NotCalledWith asserts that the mock function object was not called with a set of arguments
// matching the given call instance asserter.
func NotCalledWith(t assert.TestingT, mockFn interface{}, asserter CallInstanceAsserter, msgAndArgs ...interface{}) bool {
	matchingCallCount, ok := callCountWith(t, mockFn, asserter, msgAndArgs...)
	if !ok {
		return false
	}
	if matchingCallCount != 0 {
		return assert.Fail(t, fmt.Sprintf("Did not expect %T to be called with given arguments", mockFn), msgAndArgs...)
	}
	return true
}

// CalledOnceWith asserts that the mock function object was called exactly once with a set of
// arguments matching the given call instance asserter.
func CalledOnceWith(t assert.TestingT, mockFn interface{}, asserter CallInstanceAsserter, msgAndArgs ...interface{}) bool {
	return CalledNWith(t, mockFn, 1, asserter, msgAndArgs...)
}

// CalledNWith asserts that the mock function object was called exactly n times with a set of
// arguments matching the given call instance asserter.
func CalledNWith(t assert.TestingT, mockFn interface{}, n int, asserter CallInstanceAsserter, msgAndArgs ...interface{}) bool {
	matchingCallCount, ok := callCountWith(t, mockFn, asserter, msgAndArgs...)
	if !ok {
		return false
	}
	if matchingCallCount != n {
		return assert.Fail(t, fmt.Sprintf("Expected %T to be called with given arguments exactly %d times, called %d times", mockFn, n, matchingCallCount), msgAndArgs...)
	}
	return true
}

// CalledAtNWith asserts that the mock function objects nth call was with a set of
// arguments matching the given call instance asserter.
func CalledAtNWith(t assert.TestingT, mockFn interface{}, n int, asserter CallInstanceAsserter, msgAndArgs ...interface{}) bool {
	hist, ok := testutil.GetCallHistory(mockFn)
	if !ok {
		return false
	}
	if len(hist) < n {
		return assert.Fail(t, fmt.Sprintf("Expected %T to be called at least %d times, called %d times", mockFn, n, len(hist)), msgAndArgs...)
	}

	if !asserter.Assert(hist[n]) {
		return assert.Fail(t, fmt.Sprintf("Expected call %d of %T to be with given arguments", n, mockFn), msgAndArgs...)
	}

	return true
}

// callCount returns the number of times the given mock function was called.
func callCount(t assert.TestingT, mockFn interface{}, msgAndArgs ...interface{}) (int, bool) {
	return callCountWith(t, mockFn, CallInstanceAsserterFunc(func(call interface{}) bool { return true }), msgAndArgs...)
}

// callCount returns the number of times the given mock function was called with a set of
// arguments matching the given call instance asserter.
func callCountWith(t assert.TestingT, mockFn interface{}, asserter CallInstanceAsserter, msgAndArgs ...interface{}) (int, bool) {
	matchingHistory, ok := testutil.GetCallHistoryWith(mockFn, func(call testutil.CallInstance) bool {
		// Pass in a dummy non-erroring TestingT so that any assertions done inside
		// of the asserter will not fail the enclosing test.
		return asserter.Assert(call)
	})
	if !ok {
		return 0, assert.Fail(t, fmt.Sprintf("Parameters must be a mock function description, got %T", mockFn), msgAndArgs...)
	}

	return len(matchingHistory), true
}
