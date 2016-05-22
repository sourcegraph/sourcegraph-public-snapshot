package errcode_test

import (
	"errors"
	"net/http"
	"os"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/util/errcode"
)

func TestHTTP(t *testing.T) {
	tests := []struct {
		err  error
		want int
	}{
		{os.ErrNotExist, http.StatusNotFound},
		{grpc.Errorf(codes.NotFound, ""), http.StatusNotFound},
		{grpc.Errorf(codes.OK, ""), http.StatusOK},
		{grpc.Errorf(codes.Unknown, ""), http.StatusInternalServerError},
		{nil, http.StatusOK},
		{&store.RepoNotFoundError{}, http.StatusNotFound},
		{errors.New(""), http.StatusInternalServerError},
	}
	for _, test := range tests {
		c := errcode.HTTP(test.err)
		if c != test.want {
			t.Errorf("error %q: got %d, want %d", test.err, c, test.want)
		}
	}
}

func TestGRPC(t *testing.T) {
	tests := []struct {
		err  error
		want codes.Code
	}{
		{os.ErrNotExist, codes.NotFound},
		{&errcode.HTTPErr{Status: http.StatusNotFound}, codes.NotFound},
		{&errcode.HTTPErr{Status: http.StatusInternalServerError}, codes.Unknown},
		{grpc.Errorf(codes.NotFound, ""), codes.NotFound},
		{grpc.Errorf(codes.OK, ""), codes.OK},
		{nil, codes.OK},
		{&store.RepoNotFoundError{}, codes.NotFound},
		{errors.New(""), codes.Unknown},
	}
	for _, test := range tests {
		c := errcode.GRPC(test.err)
		if c != test.want {
			t.Errorf("error %q: got %d, want %d", test.err, c, test.want)
		}
	}
}
