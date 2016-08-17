// +build !go1.7

package csrf

import (
	"net/http"

	"github.com/gorilla/context"

	"github.com/pkg/errors"
)

func contextGet(r *http.Request, key string) (interface{}, error) {
	if val, ok := context.GetOk(r, key); ok {
		return val, nil
	}

	return nil, errors.Errorf("no value exists in the context for key %q", key)
}

func contextSave(r *http.Request, key string, val interface{}) *http.Request {
	context.Set(r, key, val)
	return r
}

func contextClear(r *http.Request) {
	context.Clear(r)
}
