package client

import (
	"net/http"

	"github.com/pkg/errors"
)

type lsifError struct {
	StatusCode int
	Message    string
}

func (e *lsifError) Error() string {
	return e.Message
}

func IsNotFound(err error) bool {
	if e, ok := errors.Cause(err).(*lsifError); ok {
		return e.StatusCode == http.StatusNotFound
	}

	return false
}
