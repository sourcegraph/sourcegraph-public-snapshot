package graphql

import "testing"

func TestCursor(t *testing.T) {
	expected := "test"
	pageInfo := encodeCursor(&expected)

	if !pageInfo.HasNextPage() {
		t.Fatalf("expected next page")
	}
	if pageInfo.EndCursor() == nil {
		t.Fatalf("unexpected nil cursor")
	}

	value, err := decodeCursor(pageInfo.EndCursor())
	if err != nil {
		t.Fatalf("unexpected error decoding cursor: %s", err)
	}
	if value != expected {
		t.Errorf("unexpected decoded cursor. want=%s have=%s", expected, value)
	}
}

func TestCursorEmpty(t *testing.T) {
	pageInfo := encodeCursor(nil)

	if pageInfo.HasNextPage() {
		t.Errorf("unexpected next page")
	}
	if pageInfo.EndCursor() != nil {
		t.Errorf("unexpected encoded cursor: %s", *pageInfo.EndCursor())
	}

	value, err := decodeCursor(nil)
	if err != nil {
		t.Fatalf("unexpected error decoding cursor: %s", err)
	}
	if value != "" {
		t.Errorf("unexpected decoded cursor: %s", value)
	}
}

func TestIntCursor(t *testing.T) {
	expected := 42
	pageInfo := encodeIntCursor(toInt32(&expected))

	if !pageInfo.HasNextPage() {
		t.Fatalf("expected next page")
	}
	if pageInfo.EndCursor() == nil {
		t.Fatalf("unexpected nil cursor")
	}

	value, err := decodeIntCursor(pageInfo.EndCursor())
	if err != nil {
		t.Fatalf("unexpected error decoding cursor: %s", err)
	}
	if value != expected {
		t.Errorf("unexpected decoded cursor. want=%d have=%d", expected, value)
	}
}

func TestIntCursorEmpty(t *testing.T) {
	pageInfo := encodeIntCursor(nil)

	if pageInfo.HasNextPage() {
		t.Errorf("unexpected next page")
	}
	if pageInfo.EndCursor() != nil {
		t.Errorf("unexpected encoded cursor: %s", *pageInfo.EndCursor())
	}

	value, err := decodeIntCursor(nil)
	if err != nil {
		t.Fatalf("unexpected error decoding cursor: %s", err)
	}
	if value != 0 {
		t.Errorf("unexpected decoded cursor: %d", value)
	}
}
