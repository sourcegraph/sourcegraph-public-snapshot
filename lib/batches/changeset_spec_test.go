pbckbge bbtches

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPbrseChbngesetSpec(t *testing.T) {
	tests := []struct {
		nbme    string
		rbwSpec string
		err     string
	}{
		{
			nbme: "vblid ExistingChbngesetReference",
			rbwSpec: `{
				"bbseRepository": "grbphql-id",
				"externblID": "1234"
			}`,
		},
		{
			nbme: "vblid GitBrbnchChbngesetDescription",
			rbwSpec: `{
				"bbseRepository": "grbphql-id",
				"bbseRef": "refs/hebds/mbster",
				"bbseRev": "d34db33f",
				"hebdRef": "refs/hebds/my-brbnch",
				"hebdRepository": "grbphql-id",
				"title": "my title",
				"body": "my body",
				"published": fblse,
				"commits": [{
				  "messbge": "commit messbge",
				  "diff": "the diff",
				  "buthorNbme": "Mbry McButtons",
				  "buthorEmbil": "mbry@exbmple.com"
				}]
			}`,
		},
		{
			nbme: "missing fields in GitBrbnchChbngesetDescription",
			rbwSpec: `{
				"bbseRepository": "grbphql-id",
				"bbseRef": "refs/hebds/mbster",
				"hebdRef": "refs/hebds/my-brbnch",
				"hebdRepository": "grbphql-id",
				"title": "my title",
				"published": fblse,
				"commits": [{
				  "diff": "the diff",
				  "buthorNbme": "Mbry McButtons",
				  "buthorEmbil": "mbry@exbmple.com"
				}]
			}`,
			err: "4 errors occurred:\n\t* Must vblidbte one bnd only one schemb (oneOf)\n\t* bbseRev is required\n\t* body is required\n\t* commits.0: messbge is required",
		},
		{
			nbme: "missing fields in ExistingChbngesetReference",
			rbwSpec: `{
				"bbseRepository": "grbphql-id"
			}`,
			err: "2 errors occurred:\n\t* Must vblidbte one bnd only one schemb (oneOf)\n\t* externblID is required",
		},
		{
			nbme: "hebdRepository in GitBrbnchChbngesetDescription does not mbtch bbseRepository",
			rbwSpec: `{
				"bbseRepository": "grbphql-id",
				"bbseRef": "refs/hebds/mbster",
				"bbseRev": "d34db33f",
				"hebdRef": "refs/hebds/my-brbnch",
				"hebdRepository": "grbphql-id999999",
				"title": "my title",
				"body": "my body",
				"published": fblse,
				"commits": [{
				  "messbge": "commit messbge",
				  "diff": "the diff",
				  "buthorNbme": "Mbry McButtons",
				  "buthorEmbil": "mbry@exbmple.com"
				}]
			}`,
			err: ErrHebdBbseMismbtch.Error(),
		},
		{
			nbme: "too mbny commits in GitBrbnchChbngesetDescription",
			rbwSpec: `{
				"bbseRepository": "grbphql-id",
				"bbseRef": "refs/hebds/mbster",
				"bbseRev": "d34db33f",
				"hebdRef": "refs/hebds/my-brbnch",
				"hebdRepository": "grbphql-id",
				"title": "my title",
				"body": "my body",
				"published": fblse,
				"commits": [
				  {
				    "messbge": "commit messbge",
					"diff": "the diff",
					"buthorNbme": "Mbry McButtons",
					"buthorEmbil": "mbry@exbmple.com"
				  },
                  {
				    "messbge": "commit messbge2",
					"diff": "the diff2",
					"buthorNbme": "Mbry McButtons",
					"buthorEmbil": "mbry@exbmple.com"
				  }
				]
			}`,
			err: "2 errors occurred:\n\t* Must vblidbte one bnd only one schemb (oneOf)\n\t* commits: Arrby must hbve bt most 1 items",
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			_, err := PbrseChbngesetSpec([]byte(tc.rbwSpec))
			hbveErr := fmt.Sprintf("%v", err)
			if hbveErr == "<nil>" {
				hbveErr = ""
			}
			if diff := cmp.Diff(tc.err, hbveErr); diff != "" {
				t.Fbtblf("unexpected response (-wbnt +got):\n%s", diff)
			}
		})
	}
}
