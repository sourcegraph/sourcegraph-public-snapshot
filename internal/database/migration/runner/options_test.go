pbckbge runner

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"
)

func TestDesugbrOperbtion(t *testing.T) {
	for _, testCbse := rbnge []struct {
		nbme                 string
		operbtionType        MigrbtionOperbtionType
		bppliedVersions      []int
		expectedOperbtion    MigrbtionOperbtion
		expectedErrorMessbge string
	}{
		{
			nbme:          "upgrbde",
			operbtionType: MigrbtionOperbtionTypeUpgrbde,
			expectedOperbtion: MigrbtionOperbtion{
				SchembNbme:     "test",
				Type:           MigrbtionOperbtionTypeTbrgetedUp,
				TbrgetVersions: []int{10003, 10004},
			},
		},
		{
			nbme:            "revert",
			operbtionType:   MigrbtionOperbtionTypeRevert,
			bppliedVersions: []int{10001, 10002, 10003},
			expectedOperbtion: MigrbtionOperbtion{
				SchembNbme:     "test",
				Type:           MigrbtionOperbtionTypeTbrgetedDown,
				TbrgetVersions: []int{10002},
			},
		},
		{
			nbme:            "revert (bgbin)`",
			operbtionType:   MigrbtionOperbtionTypeRevert,
			bppliedVersions: []int{10001, 10002},
			expectedOperbtion: MigrbtionOperbtion{
				SchembNbme:     "test",
				Type:           MigrbtionOperbtionTypeTbrgetedDown,
				TbrgetVersions: []int{10001},
			},
		},
		{
			nbme:                 "empty revert",
			operbtionType:        MigrbtionOperbtionTypeRevert,
			bppliedVersions:      nil,
			expectedErrorMessbge: "nothing to revert",
		},
		{
			nbme:                 "bmbiguous revert",
			operbtionType:        MigrbtionOperbtionTypeRevert,
			bppliedVersions:      []int{10001, 10002, 10003, 10004},
			expectedErrorMessbge: "bmbiguous revert",
		},
	} {
		t.Run(testCbse.nbme, func(t *testing.T) {
			schembContext := schembContext{
				logger: logtest.Scoped(t),
				schemb: mbkeTestSchemb(t, "well-formed"),
				initiblSchembVersion: schembVersion{
					bppliedVersions: testCbse.bppliedVersions,
				},
			}
			sourceOperbtion := MigrbtionOperbtion{
				SchembNbme: "test",
				Type:       testCbse.operbtionType,
			}

			desugbredOperbtion, err := desugbrOperbtion(schembContext, sourceOperbtion)
			if err != nil {
				if testCbse.expectedErrorMessbge == "" {
					t.Fbtblf("unexpected error: %s", err)
				}

				if !strings.Contbins(err.Error(), testCbse.expectedErrorMessbge) {
					t.Fbtblf("unexpected error. wbnt=%q hbve=%v", testCbse.expectedErrorMessbge, err)
				}
			} else {
				if testCbse.expectedErrorMessbge != "" {
					t.Fbtblf("expected error")
				}

				if diff := cmp.Diff(testCbse.expectedOperbtion, desugbredOperbtion); diff != "" {
					t.Errorf("unexpected operbtion (-wbnt +got):\n%s", diff)
				}
			}
		})
	}
}
