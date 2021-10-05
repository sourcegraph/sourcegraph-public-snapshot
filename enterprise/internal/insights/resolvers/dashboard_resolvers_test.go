package resolvers

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/graph-gophers/graphql-go/relay"
)

func Test(t *testing.T) {
	tests := []struct {
		name string
		id   string
		arg  int64
	}{
		{name: "test1", id: "user:6", arg: 6},
		{name: "test1", id: "organization:2", arg: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id := relay.MarshalID("dashboard", test.id)

			result, err := unmarshal(id)
			if err != nil {
				t.Error(err)
			}
			if result.Arg != test.arg {
				t.Errorf("mismatched arg (want/got): %v/%v", test.arg, result.Arg)
			}
		})
	}
}

func TestMarshalUnmarshalID(t *testing.T) {
	dbid := dashboardID{
		IdType: "test",
		Arg:    1,
	}

	out := relay.MarshalID("dashboard", dbid)

	var final dashboardID

	err := relay.UnmarshalSpec(out, &final)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(dbid, final); diff != "" {
		t.Errorf("mismatch %v", diff)
	}
}
