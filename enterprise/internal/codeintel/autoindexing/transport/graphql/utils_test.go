package graphql

import (
	"testing"
)

func TestIndexID(t *testing.T) {
	expected := int64(42)
	value, err := unmarshalLSIFIndexGQLID(marshalLSIFIndexGQLID(expected))
	if err != nil {
		t.Fatalf("unexpected error marshalling id: %s", err)
	}
	if value != expected {
		t.Errorf("unexpected id. have=%d want=%d", expected, value)
	}
}

func TestDerefString(t *testing.T) {
	expected := "foo"

	if val := derefString(nil, expected); val != expected {
		t.Errorf("unexpected value. want=%s have=%s", expected, val)
	}
	if val := derefString(&expected, ""); val != expected {
		t.Errorf("unexpected value. want=%s have=%s", expected, val)
	}
	if val := derefString(&expected, expected); val != expected {
		t.Errorf("unexpected value. want=%s have=%s", expected, val)
	}
}
