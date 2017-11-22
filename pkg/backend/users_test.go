package backend

import (
	"testing"
)

func TestUsers_List(t *testing.T) {
	ctx := testContext()

	_, err := Users.List(ctx)
	if err == nil {
		t.Errorf("Non-admin can access endpoint")
	}
}
