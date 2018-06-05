package db

import "testing"

func TestOrgs_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	org, err := Orgs.Create(ctx, "a", nil)
	if err != nil {
		t.Fatal(err)
	}

	if count, err := Orgs.Count(ctx, OrgsListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 1; count != want {
		t.Errorf("got %d, want %d", count, want)
	}

	if err := Orgs.Delete(ctx, org.ID); err != nil {
		t.Fatal(err)
	}

	if count, err := Orgs.Count(ctx, OrgsListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 0; count != want {
		t.Errorf("got %d, want %d", count, want)
	}
}

func TestOrgs_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	org, err := Orgs.Create(ctx, "a", nil)
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
	orgs, err := Orgs.List(ctx, nil)
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
