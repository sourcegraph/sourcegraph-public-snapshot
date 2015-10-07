package vfsutil

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	"golang.org/x/tools/godoc/vfs"
)

type MockReadSeekCloser struct {
	*bytes.Reader
}

func (_ *MockReadSeekCloser) Close() error { return nil }

type MockOpener struct {
	bufs map[string][]byte
}

func (m *MockOpener) Open(name string) (vfs.ReadSeekCloser, error) {
	if buf, ok := m.bufs[name]; ok {
		return &MockReadSeekCloser{bytes.NewReader(buf)}, nil
	} else {
		return nil, fmt.Errorf("No such path: %s", name)
	}
}

func TestConcurrentRead(t *testing.T) {
	nPaths := 1000
	paths := make([]string, nPaths)
	opener := &MockOpener{map[string][]byte{}}
	for i := 0; i < nPaths; i++ {
		paths[i] = strconv.Itoa(i)
		opener.bufs[paths[i]] = []byte(paths[i])
	}
	readCh, done := ConcurrentRead(opener, paths)
	defer done.Done()
	i := 0
	for ret := range readCh {
		if ret.Error != nil {
			t.Fatalf("Unexpected error")
		}
		out := string(ret.Bytes)
		if paths[i] != out {
			t.Fatalf("Reading path %s got path %s instead", paths[i], out)
		}
		if paths[i] != ret.Path {
			t.Fatalf("%s != %s", paths[i], ret.Path)
		}
		i += 1
	}
}

func TestConcurrentRead_Cancel(t *testing.T) {
	nPaths := 100
	paths := make([]string, nPaths)
	opener := &MockOpener{map[string][]byte{}}
	for i := 0; i < nPaths; i++ {
		paths[i] = strconv.Itoa(i)
		opener.bufs[paths[i]] = []byte(paths[i])
	}

	start := nPaths / 2
	end := start + parSize
	if nPaths < nPaths {
		t.Fatalf("Need to adjust nPaths")
	}
	for cancelIdx := start; cancelIdx < end; cancelIdx++ {
		readCh, done := ConcurrentRead(opener, paths)
		defer done.Done()
		i := 0
		for ret := range readCh {
			if i == cancelIdx {
				break
			}
			if ret.Error != nil {
				t.Fatalf("Unexpected error")
			}
			out := string(ret.Bytes)
			if paths[i] != out {
				t.Fatalf("Reading path %s got path %s instead", paths[i], out)
			}
			i += 1
		}
	}

}

func TestConcurrentRead_Empty(t *testing.T) {
	paths := []string{}
	opener := &MockOpener{map[string][]byte{}}
	readCh, done := ConcurrentRead(opener, paths)
	defer done.Done()
	for range readCh {
		t.Fatalf("Expected no items")
	}
}
