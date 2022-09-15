package dbstore

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGetConfigurationPolicyByID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStoreWithoutConfigurationPolicies(t, db)
	ctx := context.Background()

	query := `
		INSERT INTO lsif_configuration_policies (
			id,
			repository_id,
			repository_patterns,
			name,
			type,
			pattern,
			retention_enabled,
			retention_duration_hours,
			retain_intermediate_commits,
			indexing_enabled,
			index_commit_max_age_hours,
			index_intermediate_commits
		) VALUES (101, 42, '{github.com/*}', 'policy 1', 'GIT_TREE', 'ab/', true, 2, false, false, 3, true)
	`
	if _, err := db.ExecContext(ctx, query); err != nil {
		t.Fatalf("unexpected error while inserting configuration policies: %s", err)
	}

	policy, ok, err := store.GetConfigurationPolicyByID(context.Background(), 101)
	if err != nil {
		t.Fatalf("unexpected error fetching configuration policy: %s", err)
	}
	if !ok {
		t.Fatalf("expected record")
	}

	d1 := time.Hour * 2
	d2 := time.Hour * 3

	repositoryID := 42
	repositoryPatterns := []string{"github.com/*"}

	expectedPolicy := ConfigurationPolicy{
		ID:                        101,
		RepositoryID:              &repositoryID,
		RepositoryPatterns:        &repositoryPatterns,
		Name:                      "policy 1",
		Type:                      GitObjectTypeTree,
		Pattern:                   "ab/",
		RetentionEnabled:          true,
		RetentionDuration:         &d1,
		RetainIntermediateCommits: false,
		IndexingEnabled:           false,
		IndexCommitMaxAge:         &d2,
		IndexIntermediateCommits:  true,
	}
	if diff := cmp.Diff(expectedPolicy, policy); diff != "" {
		t.Errorf("unexpected configuration policy (-want +got):\n%s", diff)
	}

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Enable permissions user mapping forces checking repository permissions
		// against permissions tables in the database, which should effectively block
		// all access because permissions tables are empty.
		before := globals.PermissionsUserMapping()
		globals.SetPermissionsUserMapping(&schema.PermissionsUserMapping{Enabled: true})
		defer globals.SetPermissionsUserMapping(before)

		_, ok, err := store.GetConfigurationPolicyByID(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error fetching configuration policy: %s", err)
		}
		if ok {
			t.Fatalf("unexpected record")
		}
	})
}

func TestGetConfigurationPolicyByIDUnknownID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStoreWithoutConfigurationPolicies(t, db)

	_, ok, err := store.GetConfigurationPolicyByID(context.Background(), 15)
	if err != nil {
		t.Fatalf("unexpected error fetching configuration policy: %s", err)
	}
	if ok {
		t.Fatalf("unexpected record")
	}
}

func TestCreateConfigurationPolicy(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStoreWithoutConfigurationPolicies(t, db)

	repositoryID := 42
	d1 := time.Hour * 5
	d2 := time.Hour * 6

	configurationPolicy := ConfigurationPolicy{
		RepositoryID:              &repositoryID,
		Name:                      "name",
		Type:                      GitObjectTypeCommit,
		RepositoryPatterns:        &[]string{"a/", "b/"},
		Pattern:                   "deadbeef",
		RetentionEnabled:          false,
		RetentionDuration:         &d1,
		RetainIntermediateCommits: true,
		IndexingEnabled:           false,
		IndexCommitMaxAge:         &d2,
		IndexIntermediateCommits:  true,
	}

	hydratedConfigurationPolicy, err := store.CreateConfigurationPolicy(context.Background(), configurationPolicy)
	if err != nil {
		t.Fatalf("unexpected error creating configuration policy: %s", err)
	}

	// Inherit auto-generated identifier
	if hydratedConfigurationPolicy.ID == 0 {
		t.Fatalf("hydrated policy does not have an identifier")
	}
	configurationPolicy.ID = hydratedConfigurationPolicy.ID

	if diff := cmp.Diff(configurationPolicy, hydratedConfigurationPolicy); diff != "" {
		t.Errorf("unexpected configuration policy (-want +got):\n%s", diff)
	}

	roundTrippedConfigurationPolicy, _, err := store.GetConfigurationPolicyByID(context.Background(), configurationPolicy.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching configuration policy: %s", err)
	}

	if diff := cmp.Diff(roundTrippedConfigurationPolicy, hydratedConfigurationPolicy); diff != "" {
		t.Errorf("unexpected configuration policy (-want +got):\n%s", diff)
	}
}

func TestUpdateConfigurationPolicy(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStoreWithoutConfigurationPolicies(t, db)

	repositoryID := 42
	d1 := time.Hour * 5
	d2 := time.Hour * 6

	configurationPolicy := ConfigurationPolicy{
		RepositoryID:              &repositoryID,
		Name:                      "name",
		Type:                      GitObjectTypeCommit,
		Pattern:                   "deadbeef",
		RetentionEnabled:          false,
		RetentionDuration:         &d1,
		RetainIntermediateCommits: true,
		IndexingEnabled:           false,
		IndexCommitMaxAge:         &d2,
		IndexIntermediateCommits:  true,
	}

	hydratedConfigurationPolicy, err := store.CreateConfigurationPolicy(context.Background(), configurationPolicy)
	if err != nil {
		t.Fatalf("unexpected error creating configuration policy: %s", err)
	}

	// Inherit auto-generated identifier
	if hydratedConfigurationPolicy.ID == 0 {
		t.Fatalf("hydrated policy does not have an identifier")
	}

	d3 := time.Hour * 10
	d4 := time.Hour * 15

	newConfigurationPolicy := ConfigurationPolicy{
		ID:                        hydratedConfigurationPolicy.ID,
		RepositoryID:              &repositoryID,
		Name:                      "new name",
		Type:                      GitObjectTypeTree,
		Pattern:                   "az/",
		RetentionEnabled:          true,
		RetentionDuration:         &d3,
		RetainIntermediateCommits: false,
		IndexingEnabled:           true,
		IndexCommitMaxAge:         &d4,
		IndexIntermediateCommits:  false,
	}

	if err := store.UpdateConfigurationPolicy(context.Background(), newConfigurationPolicy); err != nil {
		t.Fatalf("unexpected error updating configuration policy: %s", err)
	}

	roundTrippedConfigurationPolicy, _, err := store.GetConfigurationPolicyByID(context.Background(), newConfigurationPolicy.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching configuration policy: %s", err)
	}

	if diff := cmp.Diff(roundTrippedConfigurationPolicy, newConfigurationPolicy); diff != "" {
		t.Errorf("unexpected configuration policy (-want +got):\n%s", diff)
	}
}

func TestUpdateProtectedConfigurationPolicy(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStoreWithoutConfigurationPolicies(t, db)

	repositoryID := 42
	d1 := time.Hour * 5
	d2 := time.Hour * 6

	configurationPolicy := ConfigurationPolicy{
		RepositoryID:              &repositoryID,
		Name:                      "default branch policy",
		Type:                      GitObjectTypeTree,
		Pattern:                   "*",
		RetentionEnabled:          true,
		RetentionDuration:         &d1,
		RetainIntermediateCommits: false,
		IndexingEnabled:           false,
		IndexCommitMaxAge:         &d2,
		IndexIntermediateCommits:  true,
	}

	hydratedConfigurationPolicy, err := store.CreateConfigurationPolicy(context.Background(), configurationPolicy)
	if err != nil {
		t.Fatalf("unexpected error creating configuration policy: %s", err)
	}

	// Inherit auto-generated identifier
	if hydratedConfigurationPolicy.ID == 0 {
		t.Fatalf("hydrated policy does not have an identifier")
	}

	// Mark configuration policy as protected (no other way to do so outside of migrations)
	if _, err := db.ExecContext(context.Background(), "UPDATE lsif_configuration_policies SET protected = true"); err != nil {
		t.Fatalf("unexpected error marking configuration policy as protected: %s", err)
	}

	t.Run("illegal update", func(t *testing.T) {
		t.Run("name", func(t *testing.T) {
			newConfigurationPolicy := hydratedConfigurationPolicy
			newConfigurationPolicy.Name = "some clever name"

			if err := store.UpdateConfigurationPolicy(context.Background(), newConfigurationPolicy); err == nil {
				t.Fatalf("expected error updating protected configuration policy")
			}
		})

		t.Run("type", func(t *testing.T) {
			newConfigurationPolicy := hydratedConfigurationPolicy
			newConfigurationPolicy.Type = GitObjectTypeTag

			if err := store.UpdateConfigurationPolicy(context.Background(), newConfigurationPolicy); err == nil {
				t.Fatalf("expected error updating protected configuration policy")
			}
		})

		t.Run("pattern", func(t *testing.T) {
			newConfigurationPolicy := hydratedConfigurationPolicy
			newConfigurationPolicy.Pattern = "ef/"

			if err := store.UpdateConfigurationPolicy(context.Background(), newConfigurationPolicy); err == nil {
				t.Fatalf("expected error updating protected configuration policy")
			}
		})

		t.Run("retentionEnabled", func(t *testing.T) {
			newConfigurationPolicy := hydratedConfigurationPolicy
			newConfigurationPolicy.RetentionEnabled = false

			if err := store.UpdateConfigurationPolicy(context.Background(), newConfigurationPolicy); err == nil {
				t.Fatalf("expected error updating protected configuration policy")
			}
		})

		t.Run("retainIntermediateCommits", func(t *testing.T) {
			newConfigurationPolicy := hydratedConfigurationPolicy
			newConfigurationPolicy.RetainIntermediateCommits = true

			if err := store.UpdateConfigurationPolicy(context.Background(), newConfigurationPolicy); err == nil {
				t.Fatalf("expected error updating protected configuration policy")
			}
		})
	})

	t.Run("success", func(t *testing.T) {
		d3 := time.Hour * 10
		d4 := time.Hour * 15

		newConfigurationPolicy := hydratedConfigurationPolicy
		newConfigurationPolicy.Protected = true
		newConfigurationPolicy.RetentionDuration = &d3
		newConfigurationPolicy.IndexingEnabled = true
		newConfigurationPolicy.IndexCommitMaxAge = &d4
		newConfigurationPolicy.IndexIntermediateCommits = false

		if err := store.UpdateConfigurationPolicy(context.Background(), newConfigurationPolicy); err != nil {
			t.Fatalf("unexpected error updating configuration policy: %s", err)
		}

		roundTrippedConfigurationPolicy, _, err := store.GetConfigurationPolicyByID(context.Background(), newConfigurationPolicy.ID)
		if err != nil {
			t.Fatalf("unexpected error fetching configuration policy: %s", err)
		}

		if diff := cmp.Diff(roundTrippedConfigurationPolicy, newConfigurationPolicy); diff != "" {
			t.Errorf("unexpected configuration policy (-want +got):\n%s", diff)
		}
	})
}

func TestDeleteConfigurationPolicyByID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStoreWithoutConfigurationPolicies(t, db)

	repositoryID := 42
	d1 := time.Hour * 5
	d2 := time.Hour * 6

	configurationPolicy := ConfigurationPolicy{
		RepositoryID:              &repositoryID,
		Name:                      "name",
		Type:                      GitObjectTypeCommit,
		Pattern:                   "deadbeef",
		RetentionEnabled:          false,
		RetentionDuration:         &d1,
		RetainIntermediateCommits: true,
		IndexingEnabled:           false,
		IndexCommitMaxAge:         &d2,
		IndexIntermediateCommits:  true,
	}

	hydratedConfigurationPolicy, err := store.CreateConfigurationPolicy(context.Background(), configurationPolicy)
	if err != nil {
		t.Fatalf("unexpected error creating configuration policy: %s", err)
	}

	if hydratedConfigurationPolicy.ID == 0 {
		t.Fatalf("hydrated policy does not have an identifier")
	}

	if err := store.DeleteConfigurationPolicyByID(context.Background(), hydratedConfigurationPolicy.ID); err != nil {
		t.Fatalf("unexpected error deleting configuration policy: %s", err)
	}

	_, ok, err := store.GetConfigurationPolicyByID(context.Background(), hydratedConfigurationPolicy.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching configuration policy: %s", err)
	}
	if ok {
		t.Fatalf("unexpected record")
	}
}

func TestDeleteConfigurationProtectedPolicy(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStoreWithoutConfigurationPolicies(t, db)

	repositoryID := 42
	d1 := time.Hour * 5
	d2 := time.Hour * 6

	configurationPolicy := ConfigurationPolicy{
		RepositoryID:              &repositoryID,
		Name:                      "name",
		Type:                      GitObjectTypeCommit,
		Pattern:                   "deadbeef",
		RetentionEnabled:          false,
		RetentionDuration:         &d1,
		RetainIntermediateCommits: true,
		IndexingEnabled:           false,
		IndexCommitMaxAge:         &d2,
		IndexIntermediateCommits:  true,
	}

	hydratedConfigurationPolicy, err := store.CreateConfigurationPolicy(context.Background(), configurationPolicy)
	if err != nil {
		t.Fatalf("unexpected error creating configuration policy: %s", err)
	}

	if hydratedConfigurationPolicy.ID == 0 {
		t.Fatalf("hydrated policy does not have an identifier")
	}

	// Mark configuration policy as protected (no other way to do so outside of migrations)
	if _, err := db.ExecContext(context.Background(), "UPDATE lsif_configuration_policies SET protected = true"); err != nil {
		t.Fatalf("unexpected error marking configuration policy as protected: %s", err)
	}

	if err := store.DeleteConfigurationPolicyByID(context.Background(), hydratedConfigurationPolicy.ID); err == nil {
		t.Fatalf("expected error deleting configuration policy: %s", err)
	}

	_, ok, err := store.GetConfigurationPolicyByID(context.Background(), hydratedConfigurationPolicy.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching configuration policy: %s", err)
	}
	if !ok {
		t.Fatalf("expected record")
	}
}

func TestSelectPoliciesForRepositoryMembershipUpdate(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStoreWithoutConfigurationPolicies(t, db)
	ctx := context.Background()

	query := `
		INSERT INTO lsif_configuration_policies (
			id,
			repository_id,
			name,
			type,
			pattern,
			repository_patterns,
			retention_enabled,
			retention_duration_hours,
			retain_intermediate_commits,
			indexing_enabled,
			index_commit_max_age_hours,
			index_intermediate_commits
		) VALUES
			(101, NULL, 'policy 1', 'GIT_TREE', 'ab/', null, true,  1, true,  true,  1, true),
			(102, NULL, 'policy 2', 'GIT_TREE', 'cd/', null, false, 2, true,  true,  2, true),
			(103, NULL, 'policy 3', 'GIT_TREE', 'ef/', null, true,  3, false, false, 3, false),
			(104, NULL, 'policy 4', 'GIT_TREE', 'gh/', null, false, 4, false, false, 4, false)
	`
	if _, err := db.ExecContext(ctx, query); err != nil {
		t.Fatalf("unexpected error while inserting configuration policies: %s", err)
	}

	sortedIDs := func(policies []ConfigurationPolicy) (ids []int) {
		for _, policy := range policies {
			ids = append(ids, policy.ID)
		}

		sort.Slice(ids, func(i, j int) bool {
			return ids[i] < ids[j]
		})
		return ids
	}

	// Can return nulls
	if policies, err := store.SelectPoliciesForRepositoryMembershipUpdate(context.Background(), 2); err != nil {
		t.Fatalf("unexpected error fetching configuration policies for repository membership update: %s", err)
	} else if diff := cmp.Diff([]int{101, 102}, sortedIDs(policies)); diff != "" {
		t.Fatalf("unexpected configuration policy list (-want +got):\n%s", diff)
	}

	// Returns new batch
	if policies, err := store.SelectPoliciesForRepositoryMembershipUpdate(context.Background(), 2); err != nil {
		t.Fatalf("unexpected error fetching configuration policies for repository membership update: %s", err)
	} else if diff := cmp.Diff([]int{103, 104}, sortedIDs(policies)); diff != "" {
		t.Fatalf("unexpected configuration policy list (-want +got):\n%s", diff)
	}

	// Recycles policies by age
	if policies, err := store.SelectPoliciesForRepositoryMembershipUpdate(context.Background(), 3); err != nil {
		t.Fatalf("unexpected error fetching configuration policies for repository membership update: %s", err)
	} else if diff := cmp.Diff([]int{101, 102, 103}, sortedIDs(policies)); diff != "" {
		t.Fatalf("unexpected configuration policy list (-want +got):\n%s", diff)
	}

	// Recycles policies by age
	if policies, err := store.SelectPoliciesForRepositoryMembershipUpdate(context.Background(), 3); err != nil {
		t.Fatalf("unexpected error fetching configuration policies for repository membership update: %s", err)
	} else if diff := cmp.Diff([]int{101, 102, 104}, sortedIDs(policies)); diff != "" {
		t.Fatalf("unexpected configuration policy list (-want +got):\n%s", diff)
	}
}
