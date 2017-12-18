package backend

import (
	"testing"
)

func TestOrgs_List(t *testing.T) {
	ctx := testContext()

	_, err := Orgs.List(ctx)
	if err == nil {
		t.Errorf("Non-admin can access endpoint")
	}
}
