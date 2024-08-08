package search

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"testing/quick"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func TestNewFilter(t *testing.T) {
	gitserverClient := gitserver.NewMockClient()
	gitserverClient.NewFileReaderFunc.SetDefaultReturn(io.NopCloser(bytes.NewReader([]byte("foo/"))), nil)

	ig, err := NewFilterFactory(gitserverClient)(context.Background(), "", "")
	if err != nil {
		t.Error(err)
	}

	cases := []struct {
		tar.Header
		Ignore bool
	}{{
		Ignore: true,
		Header: tar.Header{
			Name: "foo/ignore-me.go",
		},
	}, {
		Ignore: false,
		Header: tar.Header{
			Name: "bar/dont-ignore-me.go",
		},
	}, {
		// https://github.com/sourcegraph/sourcegraph/issues/23841
		Ignore: true,
		Header: tar.Header{
			Name: "bar/large-file.go",
			Size: 2 << 21,
		},
	}}

	for _, tc := range cases {
		got := ig(&tc.Header)
		if got != tc.Ignore {
			t.Errorf("unexpected ignore want=%v got %v for %v", tc.Ignore, got, tc.Header.Name)
		}
	}
}

func TestMissingIgnoreFile(t *testing.T) {
	gitserverClient := gitserver.NewMockClient()
	gitserverClient.NewFileReaderFunc.SetDefaultReturn(nil, os.ErrNotExist)

	ig, err := NewFilterFactory(gitserverClient)(context.Background(), "", "")
	if err != nil {
		t.Error(err)
	}

	// Quick check that we don't ignore.
	f := func(name string) bool {
		return !ig(&tar.Header{
			Name: name,
		})
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
