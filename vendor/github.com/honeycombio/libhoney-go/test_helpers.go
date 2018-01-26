package libhoney

import (
	"fmt"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func testOK(t testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("%s:%d: unexpected error: %s", filepath.Base(file), line, err.Error())
	}
}
func testErr(t testing.TB, err error) {
	if err == nil {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("%s:%d: error expected!", filepath.Base(file), line)
	}
}

func testEquals(t testing.TB, actual, expected interface{}, msg ...string) {
	if !reflect.DeepEqual(actual, expected) {
		testCommonErr(t, actual, expected, msg)
	}
}

func testNotEquals(t testing.TB, actual, expected interface{}, msg ...string) {
	if reflect.DeepEqual(actual, expected) {
		testCommonErr(t, actual, expected, msg)
	}
}

func testCommonErr(t testing.TB, actual, expected interface{}, msg []string) {
	message := strings.Join(msg, ", ")
	_, file, line, _ := runtime.Caller(2)

	t.Errorf(
		"%s:%d: %s -- actual(%T): %v, expected(%T): %v",
		filepath.Base(file),
		line,
		message,
		testDeref(actual),
		testDeref(actual),
		testDeref(expected),
		testDeref(expected),
	)
}

func testGetResponse(t testing.TB, ch chan Response) Response {
	_, file, line, _ := runtime.Caller(2)
	var resp Response
	select {
	case resp = <-ch:
	case <-time.After(50 * time.Millisecond): // block on read but prevent deadlocking tests
		t.Errorf("%s:%d: expected response on channel and timed out waiting for it!", filepath.Base(file), line)
	}
	return resp
}

func testIsPlaceholderResponse(t testing.TB, actual Response, msg ...string) {
	if actual.StatusCode != http.StatusTeapot {
		message := strings.Join(msg, ", ")
		_, file, line, _ := runtime.Caller(1)
		t.Errorf(
			"%s:%d placeholder expected -- %s",
			filepath.Base(file),
			line,
			message,
		)
	}
}

func testDeref(v interface{}) interface{} {
	switch t := v.(type) {
	case *string:
		return fmt.Sprintf("*(%v)", *t)
	case *int64:
		return fmt.Sprintf("*(%v)", *t)
	case *float64:
		return fmt.Sprintf("*(%v)", *t)
	case *bool:
		return fmt.Sprintf("*(%v)", *t)
	default:
		return v
	}
}

// for easy time manipulation during tests
type fakeNower struct {
	iter int
}

// Now() supports changing/increasing the returned Now() based on the number of
// times it's called in succession
func (f *fakeNower) Now() time.Time {
	now := time.Unix(1277132645, 0).Add(time.Second * 10 * time.Duration(f.iter))
	f.iter++
	return now
}
