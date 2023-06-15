package workerutil

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIDAddRemove(t *testing.T) {
	var called1, called2, called3 bool

	idSet := newIDSet()
	if !idSet.Add("1", func() { called1 = true }) {
		t.Fatalf("expected add to succeed")
	}
	if !idSet.Add("2", func() { called2 = true }) {
		t.Fatalf("expected add to succeed")
	}
	if idSet.Add("1", func() { called3 = true }) {
		t.Fatalf("expected duplicate add to fail")
	}

	idSet.Remove("1")

	if !called1 {
		t.Fatalf("expected first function to be called")
	}
	if called2 {
		t.Fatalf("did not expect second function to be called")
	}
	if called3 {
		t.Fatalf("did not expect third function to be called")
	}

	if diff := cmp.Diff([]string{"2"}, idSet.Slice()); diff != "" {
		t.Errorf("unexpected slice (-want +got):\n%s", diff)
	}
}

func TestIDSetSlice(t *testing.T) {
	idSet := newIDSet()
	idSet.Add("2", nil)
	idSet.Add("4", nil)
	idSet.Add("5", nil)
	idSet.Add("1", nil)
	idSet.Add("3", nil)

	if diff := cmp.Diff([]string{"1", "2", "3", "4", "5"}, idSet.Slice()); diff != "" {
		t.Errorf("unexpected slice (-want +got):\n%s", diff)
	}
}
