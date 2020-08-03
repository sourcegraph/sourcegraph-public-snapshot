package graphql

import "testing"

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

func TestDerefInt32(t *testing.T) {
	expected := 42
	expected32 := int32(expected)

	if val := derefInt32(nil, expected); val != expected {
		t.Errorf("unexpected value. want=%d have=%d", expected, val)
	}
	if val := derefInt32(&expected32, expected); val != expected {
		t.Errorf("unexpected value. want=%d have=%d", expected, val)
	}
}

func TestDerefBool(t *testing.T) {
	if val := derefBool(nil, true); !val {
		t.Errorf("unexpected value. want=%v have=%v", true, val)
	}
	if val := derefBool(nil, false); val {
		t.Errorf("unexpected value. want=%v have=%v", false, val)
	}

	pVal := true
	if val := derefBool(&pVal, true); !val {
		t.Errorf("unexpected value. want=%v have=%v", true, val)
	}
	if val := derefBool(&pVal, false); !val {
		t.Errorf("unexpected value. want=%v have=%v", false, val)
	}

	pVal = false
	if val := derefBool(&pVal, true); val {
		t.Errorf("unexpected value. want=%v have=%v", true, val)
	}
	if val := derefBool(&pVal, false); val {
		t.Errorf("unexpected value. want=%v have=%v", false, val)
	}
}
