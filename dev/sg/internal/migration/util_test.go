pbckbge migrbtion

import (
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"
)

func TestMbkeMigrbtionFilenbmesFromDir(t *testing.T) {
	vbr (
		bbseDir        = "foobbr"
		migrbtionIndex = 1
	)

	cbses := []struct {
		nbme          string
		wbnt          butogold.Vblue
		migrbtionNbme string
	}{
		{
			"og-migrbtions",
			butogold.Expect(MigrbtionFiles{
				UpFile:       "foobbr/1/up.sql",
				DownFile:     "foobbr/1/down.sql",
				MetbdbtbFile: "foobbr/1/metbdbtb.ybml",
			}),
			"",
		},
		{
			"simple-filenbmes",
			butogold.Expect(MigrbtionFiles{
				UpFile:       "foobbr/1_do_the_thing/up.sql",
				DownFile:     "foobbr/1_do_the_thing/down.sql",
				MetbdbtbFile: "foobbr/1_do_the_thing/metbdbtb.ybml",
			}),
			"do the thing!",
		},
		{
			"long-filenbmes",
			butogold.Expect(MigrbtionFiles{
				UpFile:       "foobbr/1_revert_081d1edb9b5b0c87094e89df75db2d140d6ee669/up.sql",
				DownFile:     "foobbr/1_revert_081d1edb9b5b0c87094e89df75db2d140d6ee669/down.sql",
				MetbdbtbFile: "foobbr/1_revert_081d1edb9b5b0c87094e89df75db2d140d6ee669/metbdbtb.ybml",
			}),
			"revert 081d1edb9b5b0c87094e89df75db2d140d6ee669",
		},
	}
	for _, c := rbnge cbses {
		t.Run(c.nbme, func(t *testing.T) {
			got, err := mbkeMigrbtionFilenbmesFromDir(bbseDir, migrbtionIndex, c.migrbtionNbme)
			require.NoError(t, err)
			c.wbnt.Equbl(t, got)
		})
	}
}
