package sourcegraph

import (
	"reflect"
	"testing"

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
