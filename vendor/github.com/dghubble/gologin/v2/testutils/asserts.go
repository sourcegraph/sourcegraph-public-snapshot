// Package testutils provides utilities for writing gologin tests.
package testutils

import (
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertSuccessNotCalled is a success http.Handler that fails if called.
func AssertSuccessNotCalled(t *testing.T) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "unexpected call to success Handler")
	}
	return http.HandlerFunc(fn)
}

// AssertFailureNotCalled is a failure http.Handler that fails if called.
func AssertFailureNotCalled(t *testing.T) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "unexpected call to failure Handler")
	}
	return http.HandlerFunc(fn)
}

// AssertBodyString asserts that a Request Body matches the expected string.
func AssertBodyString(t *testing.T, rc io.ReadCloser, expected string) {
	defer rc.Close()
	if b, err := ioutil.ReadAll(rc); err == nil {
		if string(b) != expected {
			t.Errorf("expected %q, got %q", expected, string(b))
		}
	} else {
		t.Errorf("error reading Body")
	}
}
