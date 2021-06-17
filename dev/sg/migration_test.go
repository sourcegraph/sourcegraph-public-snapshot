package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSameMigrations(t *testing.T) {
	branch := map[int]migration{
		1: {
			ID:   1,
			Name: "first",
		},
	}

	if conflicts, _, _ := findConflictingMigrations(branch, branch); len(conflicts) > 0 {
		t.Errorf("Failed when comparing exacly the same migrations: %+v", branch)
	}
}

func TestBranchHasExtraMigration(t *testing.T) {
	trunk := map[int]migration{
		1: {
			ID:   1,
			Name: "first",
		},
	}

	branch := map[int]migration{
		1: {
			ID:   1,
			Name: "first",
		},
		2: {
			ID:   2,
			Name: "second",
		},
	}

	if conflicts, _, _ := findConflictingMigrations(trunk, branch); len(conflicts) > 0 {
		t.Errorf("Failed when comparing exacly the same migrations: %+v", branch)
	}
}

func TestTrunkHasExtraMigration(t *testing.T) {
	trunk := map[int]migration{
		1: {
			ID:   1,
			Name: "first",
		},
		2: {
			ID:   2,
			Name: "second",
		},
	}

	branch := map[int]migration{
		1: {
			ID:   1,
			Name: "first",
		},
	}

	_, missing, err := findConflictingMigrations(trunk, branch)
	if err != nil {
		t.Errorf("Should not error for missing a migration.")
	}

	if len(missing) != 1 || missing[0].Name != "second" {
		t.Errorf("Should have returned only 'second' was missing, instead +%v", missing)
	}
}

func TestTrunkHasConflictingMigration(t *testing.T) {
	trunk := map[int]migration{
		1: {
			ID:   1,
			Name: "first",
		},
		2: {
			ID:   2,
			Name: "trunk second",
		},
	}

	branch := map[int]migration{
		1: {
			ID:   1,
			Name: "first",
		},
		2: {
			ID:   2,
			Name: "branch second",
		},
	}

	conflicts, _, _ := findConflictingMigrations(trunk, branch)
	expected := []migrationConflict{
		{
			ID:    2,
			Main:  trunk[2],
			Local: branch[2],
		},
	}

	if diff := cmp.Diff(conflicts, expected); diff != "" {
		t.Errorf("Not conflicts expected: %s", diff)
	}
}
