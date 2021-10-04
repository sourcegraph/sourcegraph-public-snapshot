package resolvers

import (
	"fmt"
	"testing"

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
			if result.arg != test.arg {
				t.Errorf("mismatched arg (want/got): %v/%v", test.arg, result.arg)
			}
		})
	}
}

func TestA(t *testing.T) {
	out := relay.MarshalID("dashboard", "real:6")
	fmt.Println(out)
}
