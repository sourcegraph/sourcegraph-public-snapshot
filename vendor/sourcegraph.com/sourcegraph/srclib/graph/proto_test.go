package graph

import (
	"log"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/srclib/ann"

	"github.com/gogo/protobuf/proto"
)

func TestProtobufMarshal(t *testing.T) {
	o := Output{
		Defs: []*Def{{File: "f1"}},
		Refs: []*Ref{{File: "f2"}},
		Docs: []*Doc{{File: "f3"}},
		Anns: []*ann.Ann{{Unit: "foo"}},
	}

	b, err := proto.Marshal(&o)
	if err != nil {
		t.Fatal(err)
	}

	var o2 Output
	if err := proto.Unmarshal(b, &o2); err != nil {
		log.Fatal(err)
	}

	if !reflect.DeepEqual(o2, o) {
		t.Errorf("got %#v, want %#v", o2, o)
	}
}
