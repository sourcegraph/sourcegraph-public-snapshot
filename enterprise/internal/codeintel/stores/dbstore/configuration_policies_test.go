package dbstore

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
)

func TestGetConfigurationPolicies(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	store := testStore(db)

	query := sqlf.Sprintf(`
		INSERT INTO lsif_configuration_policies (
			id,
			repository_id,
			name,
			type,
			pattern,
			retention_enabled,
			retention_duration_hours,
			retain_intermediate_commits,
			indexing_enabled,
			index_commit_max_age_hours,
			index_intermediate_commits
		) VALUES
			(1, 42,   'policy 1', 'GIT_TREE',   'ab/',      true,  2, false, false, 3, true),
			(2, 42,   'policy 2', 'GIT_TREE',   'nm/',      false, 3, true,  false, 4, false),
			(3, 43,   'policy 3', 'GIT_TREE',   'xy/',      true,  4, false, true,  5, false),
			(4, NULL, 'policy 4', 'GIT_COMMIT', 'deadbeef', false, 5, true,  false, 6, true),
			(5, NULL, 'policy 5', 'GIT_TAG',    '3.0',      false, 6, false, true,  6, false)
	`)

	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error while inserting configuration policies: %s", err)
	}

	t.Run("Global", func(t *testing.T) {
		policies, err := store.GetConfigurationPolicies(context.Background(), GetConfigurationPoliciesOptions{})
		if err != nil {
			t.Fatalf("unexpected error fetching configuration policies: %s", err)
		}

		d1 := time.Hour * 5
		d2 := time.Hour * 6

		expected := []ConfigurationPolicy{
			{
				ID:                        4,
				RepositoryID:              nil,
				Name:                      "policy 4",
				Type:                      GitObjectTypeCommit,
				Pattern:                   "deadbeef",
				RetentionEnabled:          false,
				RetentionDuration:         &d1,
				RetainIntermediateCommits: true,
				IndexingEnabled:           false,
				IndexCommitMaxAge:         &d2,
				IndexIntermediateCommits:  true,
			},
			{
				ID:                        5,
				RepositoryID:              nil,
				Name:                      "policy 5",
				Type:                      GitObjectTypeTag,
				Pattern:                   "3.0",
				RetentionEnabled:          false,
				RetentionDuration:         &d2,
				RetainIntermediateCommits: false,
				IndexingEnabled:           true,
				IndexCommitMaxAge:         &d2,
				IndexIntermediateCommits:  false,
			},
		}
		if diff := cmp.Diff(expected, policies); diff != "" {
			t.Errorf("unexpected configuration policies (-want +got):\n%s", diff)
		}
	})

	t.Run("Repository", func(t *testing.T) {
		repositoryID := 42

		policies, err := store.GetConfigurationPolicies(context.Background(), GetConfigurationPoliciesOptions{
			RepositoryID: repositoryID,
		})
		if err != nil {
			t.Fatalf("unexpected error fetching configuration policies: %s", err)
		}

		d1 := time.Hour * 2
		d2 := time.Hour * 3
		d3 := time.Hour * 3
		d4 := time.Hour * 4

		expected := []ConfigurationPolicy{
			{
				ID:                        1,
				RepositoryID:              &repositoryID,
				Name:                      "policy 1",
				Type:                      GitObjectTypeTree,
				Pattern:                   "ab/",
				RetentionEnabled:          true,
				RetentionDuration:         &d1,
				RetainIntermediateCommits: false,
				IndexingEnabled:           false,
				IndexCommitMaxAge:         &d2,
				IndexIntermediateCommits:  true,
			},
			{
				ID:                        2,
				RepositoryID:              &repositoryID,
				Name:                      "policy 2",
				Type:                      GitObjectTypeTree,
				Pattern:                   "nm/",
				RetentionEnabled:          false,
				RetentionDuration:         &d3,
				RetainIntermediateCommits: true,
				IndexingEnabled:           false,
				IndexCommitMaxAge:         &d4,
				IndexIntermediateCommits:  false,
			},
		}
		if diff := cmp.Diff(expected, policies); diff != "" {
			t.Errorf("unexpected configuration policies (-want +got):\n%s", diff)
		}
	})
}

func TestGetConfigurationPolicyByIDUnknownID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	store := testStore(db)

	_, ok, err := store.GetConfigurationPolicyByID(context.Background(), 15)
	if err != nil {
		t.Fatalf("unexpected error fetching configuration policy: %s", err)
	}
	if ok {
		t.Fatalf("unexpected record")
	}
}

func TestCreateConfigurationPolicy(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	store := testStore(db)

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
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	store := testStore(db)

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
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	store := testStore(db)

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
	if _, err := db.Exec("UPDATE lsif_configuration_policies SET protected = true"); err != nil {
		t.Fatalf("unexpected error marking configuration policy as protected: %s", err)
	}

	t.Run("illegal update", func(t *testing.T) {
		t.Run("Name", func(t *testing.T) {
			newConfigurationPolicy := hydratedConfigurationPolicy
			newConfigurationPolicy.Name = "some clever name"

			if err := store.UpdateConfigurationPolicy(context.Background(), newConfigurationPolicy); err == nil {
				t.Fatalf("expected error updating protected configuration policy")
			}
		})

		t.Run("Type", func(t *testing.T) {
			newConfigurationPolicy := hydratedConfigurationPolicy
			newConfigurationPolicy.Type = GitObjectTypeTag

			if err := store.UpdateConfigurationPolicy(context.Background(), newConfigurationPolicy); err == nil {
				t.Fatalf("expected error updating protected configuration policy")
			}
		})

		t.Run("Pattern", func(t *testing.T) {
			newConfigurationPolicy := hydratedConfigurationPolicy
			newConfigurationPolicy.Pattern = "ef/"

			if err := store.UpdateConfigurationPolicy(context.Background(), newConfigurationPolicy); err == nil {
				t.Fatalf("expected error updating protected configuration policy")
			}
		})

		t.Run("RetentionEnabled", func(t *testing.T) {
			newConfigurationPolicy := hydratedConfigurationPolicy
			newConfigurationPolicy.RetentionEnabled = false

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
		newConfigurationPolicy.RetainIntermediateCommits = true
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
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	store := testStore(db)

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
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	store := testStore(db)

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
	if _, err := db.Exec("UPDATE lsif_configuration_policies SET protected = true"); err != nil {
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
