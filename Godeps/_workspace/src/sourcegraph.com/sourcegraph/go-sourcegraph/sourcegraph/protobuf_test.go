package sourcegraph

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/srclib/graph"

	"github.com/gogo/protobuf/proto"
)

func TestProtobuf_RepoListOptions(t *testing.T) {
	v := RepoListOptions{
		Owner:       "o",
		ListOptions: ListOptions{Page: 5},
	}
	b, err := proto.Marshal(&v)
	if err != nil {
		t.Fatal(err)
	}

	var v2 RepoListOptions
	if err := proto.Unmarshal(b, &v2); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, v2) {
		t.Errorf("got %+v, want %+v", v2, v)
	}
}

func TestProtobuf_Ref(t *testing.T) {
	v := &Ref{
		Ref: graph.Ref{
			File: "f",
		},
		Authorship: &AuthorshipInfo{AuthorEmail: "a@a.com"},
	}
	b, err := proto.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	v2 := new(Ref)
	if err := proto.Unmarshal(b, v2); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, v2) {
		t.Errorf("got %+v, want %+v", v2, v)
	}
}

func TestProtobuf_Example(t *testing.T) {
	v := &Example{
		Ref: graph.Ref{
			File: "f",
		},
		SourceCode: &SourceCode{NumRefs: 123},
		StartLine:  7,
	}
	b, err := proto.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	v2 := new(Example)
	if err := proto.Unmarshal(b, v2); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v, v2) {
		t.Errorf("got %+v, want %+v", v2, v)
	}
}
