package localstore

import "testing"

func TestOrgs_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	org, err := Orgs.Create(ctx, "a", "b")
	if err != nil {
		t.Fatal(err)
	}

	// Delete org.
	if err := Orgs.Delete(ctx, org.ID); err != nil {
		t.Fatal(err)
	}

	// Org no longer exists.
	_, err = Orgs.GetByID(ctx, org.ID)
	if _, ok := err.(*OrgNotFoundError); !ok {
		t.Errorf("got error %v, want *OrgNotFoundError", err)
	}
	orgs, err := Orgs.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(orgs) > 0 {
		t.Errorf("got %d orgs, want 0", len(orgs))
	}

	// Can't delete already-deleted org.
	err = Orgs.Delete(ctx, org.ID)
	if _, ok := err.(*OrgNotFoundError); !ok {
		t.Errorf("got error %v, want *OrgNotFoundError", err)
	}
}
