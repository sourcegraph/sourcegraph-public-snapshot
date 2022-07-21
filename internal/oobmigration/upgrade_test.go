package oobmigration

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestScheduleMigrationInterrupts(t *testing.T) {
	migration := func(id, iMajor, iMinor, jMajor, jMinor int) yamlMigration {
		return yamlMigration{
			ID:                     id,
			IntroducedVersionMajor: iMajor, IntroducedVersionMinor: iMinor,
			DeprecatedVersionMajor: &jMajor, DeprecatedVersionMinor: &jMinor,
		}
	}

	for _, testCase := range []struct {
		name       string
		migrations []yamlMigration
		interrupts []MigrationInterrupt
	}{
		{
			name:       "empty",
			migrations: []yamlMigration{},
			interrupts: []MigrationInterrupt{},
		},
		{
			name: "non-overlapping",
			migrations: []yamlMigration{
				// 1: [------]
				// 2: .      . [------]
				// 3: .      . .      . [------]
				// 4: .      . .      . .      . [------]
				//    32 33 34 35 36 37 38 39 40 41 42 43
				//          **       **       **       **

				migration(1 /* introduced = */, 3, 32 /* deprecated = */, 3, 34),
				migration(2 /* introduced = */, 3, 35 /* deprecated = */, 3, 37),
				migration(3 /* introduced = */, 3, 38 /* deprecated = */, 3, 40),
				migration(4 /* introduced = */, 3, 41 /* deprecated = */, 3, 43),
			},
			interrupts: []MigrationInterrupt{
				{Version: Version{3, 34}, MigrationIDs: []int{1}},
				{Version: Version{3, 37}, MigrationIDs: []int{2}},
				{Version: Version{3, 40}, MigrationIDs: []int{3}},
				{Version: Version{3, 43}, MigrationIDs: []int{4}},
			},
		},
		{
			name: "overlapping",
			migrations: []yamlMigration{
				// 1: [------]
				// 2: .  [------]
				// 3: .  .   .  . [---------------]
				// 4: .  .   .  . .  [---]        .
				// 5: .  .   .  . .  .   . [---]  .
				// 6: .  .   .  . .  .   . .   .  . [---]
				//    .  .   .  . .  .   . .   .  . .   .
				//    32 33 34 35 36 37 38 39 40 41 42 43
				//          **          **    **       **

				migration(1 /* introduced = */, 3, 32 /* deprecated = */, 3, 34),
				migration(2 /* introduced = */, 3, 33 /* deprecated = */, 3, 35),
				migration(3 /* introduced = */, 3, 36 /* deprecated = */, 3, 41),
				migration(4 /* introduced = */, 3, 37 /* deprecated = */, 3, 38),
				migration(5 /* introduced = */, 3, 39 /* deprecated = */, 3, 40),
				migration(6 /* introduced = */, 3, 42 /* deprecated = */, 3, 43),
			},
			interrupts: []MigrationInterrupt{
				{Version: Version{3, 34}, MigrationIDs: []int{1, 2}},
				{Version: Version{3, 38}, MigrationIDs: []int{4}},
				{Version: Version{3, 40}, MigrationIDs: []int{3, 5}},
				{Version: Version{3, 43}, MigrationIDs: []int{6}},
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			interrupts, err := scheduleMigrationInterrupts(testCase.migrations)
			if err != nil {
				t.Fatalf("falied to schedule upgrade: %s", err)
			}
			if diff := cmp.Diff(testCase.interrupts, interrupts); diff != "" {
				t.Fatalf("unexpected interrupts (-want +got):\n%s", diff)
			}
		})
	}
}
