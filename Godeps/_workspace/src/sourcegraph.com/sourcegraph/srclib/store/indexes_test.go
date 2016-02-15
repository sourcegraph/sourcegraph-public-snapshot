package store

import (
	"reflect"
	"testing"
)

func TestIndexes(t *testing.T) {
	t.Skip("TODO(sqs): add test cases")

	tests := []struct {
		store interface{}
		want  []IndexStatus
	}{}
	for _, test := range tests {
		xs, err := Indexes(test.store, IndexCriteria{}, nil)
		if err != nil {
			t.Errorf("%s: Indexes: %s", test.store, err)
			continue
		}
		if !reflect.DeepEqual(xs, test.want) {
			t.Errorf("%s: got index statuses %v, want %v", test.store, xs, test.want)
		}
	}
}
