package own

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestSearchFilteringExample(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := edb.NewEnterpriseDB(database.NewDB(logger, dbtest.NewDB(logger, t)))
	ctx := context.Background()
	user, err := db.Users().Create(ctx, database.NewUser{
		Email:           "john.doe@example.com",
		Username:        "jdoe",
		EmailIsVerified: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Imagine this is search with filter `file:has.owner(jdoe)`.
	ownerSearchTerm := "jdoe"

	// Do this at first during search and hold references to all the known entities
	// that can be referred to by given search term
	bag, err := ByTextReference(ctx, db, ownerSearchTerm)
	if err != nil {
		t.Fatalf("ByTextReference: %s", err)
	}

	// Then for given file we have owner matches (translated to references here):
	ownerReferences := map[string]Reference{
		// Some possible matching entries:
		// email entry in CODEOWNERS
		"email entry in CODEOWNERS": {
			Email: "john.doe@example.com",
			RepoContext: &RepoContext{
				Name:         "github.com/sourcegraph/sourcegraph",
				CodehostKind: "github",
			},
		},
		"jdoe entry in CODEOWNERS": {
			Handle: "jdoe",
			RepoContext: &RepoContext{
				Name:         "github.com/sourcegraph/sourcegraph",
				CodehostKind: "github",
			},
		},
		"@jdoe entry in CODEOWNERS": {
			Handle: "@jdoe",
			RepoContext: &RepoContext{
				Name:         "github.com/sourcegraph/sourcegraph",
				CodehostKind: "github",
			},
		},
		"user ID from assigned ownership": {
			UserID: user.ID,
		},
	}
	for name, r := range ownerReferences {
		t.Run(name, func(t *testing.T) {
			if !bag.Contains(r) {
				t.Errorf("bag.Contains(%s), want true, got false", r)
			}
		})
	}
}

func TestBagNoUser(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)

	db := edb.NewEnterpriseDB(database.NewDB(logger, dbtest.NewDB(logger, t)))
	ctx := context.Background()
	bag, err := ByTextReference(ctx, db, "userdoesnotexist")
	if err != nil {
		t.Fatalf("ByTextReference: %s", err)
	}
	for name, r := range map[string]Reference{
		"same handle matches even when there is no user": {
			Handle: "userdoesnotexist",
			RepoContext: &RepoContext{
				Name:         "github.com/sourcegraph/sourcegraph",
				CodehostKind: "github",
			},
		},
		"same handle with @ matches even when there is no user": {
			Handle: "@userdoesnotexist",
			RepoContext: &RepoContext{
				Name:         "github.com/sourcegraph/sourcegraph",
				CodehostKind: "github",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if !bag.Contains(r) {
				t.Errorf("bag.Contains(%s), want true, got false", r)
			}
		})
	}
	for name, r := range map[string]Reference{
		"email entry in CODEOWNERS": {
			Email: "john.doe@example.com",
			RepoContext: &RepoContext{
				Name:         "github.com/sourcegraph/sourcegraph",
				CodehostKind: "github",
			},
		},
		"different handle entry in CODEOWNERS": {
			Handle: "anotherhandle",
			RepoContext: &RepoContext{
				Name:         "github.com/sourcegraph/sourcegraph",
				CodehostKind: "github",
			},
		},
		"user ID from assigned ownership": {
			UserID: 42,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if bag.Contains(r) {
				t.Errorf("bag.Contains(%s), want false, got true", r)
			}
		})
	}
}

func TestBagUserFoundNoMatches(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := edb.NewEnterpriseDB(database.NewDB(logger, dbtest.NewDB(logger, t)))
	ctx := context.Background()
	user, err := db.Users().Create(ctx, database.NewUser{
		Email:           "john.doe@example.com",
		Username:        "jdoe",
		EmailIsVerified: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	bag, err := ByTextReference(ctx, db, "jdoe")
	if err != nil {
		t.Fatalf("ByTextReference: %s", err)
	}
	// Check test is valid by verifying user can be found by handle.
	r := Reference{Handle: "jdoe"}
	if !bag.Contains(r) {
		t.Fatalf("validation: Contains(%s), want true, got false", r)
	}
	for name, r := range map[string]Reference{
		"email entry in CODEOWNERS": {
			Email: "jdoe@example.com",
			RepoContext: &RepoContext{
				Name:         "github.com/sourcegraph/sourcegraph",
				CodehostKind: "github",
			},
		},
		"different handle entry in CODEOWNERS": {
			Handle: "anotherhandle",
			RepoContext: &RepoContext{
				Name:         "github.com/sourcegraph/sourcegraph",
				CodehostKind: "github",
			},
		},
		"user ID from assigned ownership": {
			UserID: user.ID + 1, // different user ID
		},
	} {
		t.Run(name, func(t *testing.T) {
			if bag.Contains(r) {
				t.Errorf("bag.Contains(%s), want false, got true", r)
			}
		})
	}
}
