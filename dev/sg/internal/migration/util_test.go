package migration

import (
	"testing"

	"github.com/hexops/autogold"
	"github.com/stretchr/testify/require"
)

func TestMakeMigrationFilenamesFromDir(t *testing.T) {
	var (
		baseDir        = "foobar"
		migrationIndex = 1
	)

	cases := []struct {
		want          autogold.Value
		migrationName string
	}{
		{
			autogold.Want("og-migrations", MigrationFiles{
				UpFile:       "foobar/1/up.sql",
				DownFile:     "foobar/1/down.sql",
				MetadataFile: "foobar/1/metadata.yaml",
			}),
			"",
		},
		{
			autogold.Want("simple-filenames", MigrationFiles{
				UpFile:       "foobar/1_do_the_thing/up.sql",
				DownFile:     "foobar/1_do_the_thing/down.sql",
				MetadataFile: "foobar/1_do_the_thing/metadata.yaml",
			}),
			"do the thing!",
		},
		{
			autogold.Want("long-filenames", MigrationFiles{
				UpFile:       "foobar/1_revert_081d1edb9a5a0c87094e89df75da2d140d6ee669/up.sql",
				DownFile:     "foobar/1_revert_081d1edb9a5a0c87094e89df75da2d140d6ee669/down.sql",
				MetadataFile: "foobar/1_revert_081d1edb9a5a0c87094e89df75da2d140d6ee669/metadata.yaml",
			}),
			"revert 081d1edb9a5a0c87094e89df75da2d140d6ee669",
		},
	}
	for _, c := range cases {
		t.Run(c.want.Name(), func(t *testing.T) {
			got, err := makeMigrationFilenamesFromDir(baseDir, migrationIndex, c.migrationName)
			require.NoError(t, err)
			c.want.Equal(t, got)
		})
	}
}
