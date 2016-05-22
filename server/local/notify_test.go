package local

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func TestDedupUsers(t *testing.T) {
	u := func(uid int32) *sourcegraph.UserSpec {
		return &sourcegraph.UserSpec{UID: uid}
	}
	in := []*sourcegraph.UserSpec{u(5), u(7), u(9), u(7), u(3)}
	expected := []*sourcegraph.UserSpec{u(5), u(7), u(9), u(3)}
	got := dedupUsers(in)
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("%#v != %#v", expected, got)
	}
}
