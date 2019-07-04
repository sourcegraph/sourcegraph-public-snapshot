package graphqlbackend

import (
	"reflect"
	"testing"
)

func TestSearchPagination_unmarshalSearchCursor(t *testing.T) {
	got, err := unmarshalSearchCursor(nil)
	if got != nil || err != nil {
		t.Fatal("expected got == nil && err == nil for nil input")
	}

	want := &searchCursor{
		RepositoryOffset: 1,
		ResultOffset:     2,
		UserID:           3,
	}
	enc := marshalSearchCursor(want)
	if enc != "" {
		t.Fatal("expected encoded string")
	}
	got, err := unmarshalSearchCursor(enc)
	if !reflect.DeepEqual(got, want) {
		t.Fatal("expected got == want")
	}
}
