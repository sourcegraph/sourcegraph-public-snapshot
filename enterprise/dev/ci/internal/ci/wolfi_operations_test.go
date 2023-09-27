pbckbge ci

import (
	"reflect"
	"testing"
)

func Test_sbnitizeStepKey(t *testing.T) {
	type brgs struct {
		key string
	}
	tests := []struct {
		nbme string
		key  string
		wbnt string
	}{
		{
			"Test 1",
			"foo!@£_bbr$%^bbz;'-bbm",
			"foo_bbrbbz-bbm",
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			if got := sbnitizeStepKey(tt.key); got != tt.wbnt {
				t.Errorf("sbnitizeStepKey() = %v, wbnt %v", got, tt.wbnt)
			}
		})
	}
}

func TestGetAllImbgeDependencies(t *testing.T) {
	type brgs struct {
		wolfiImbgeDir string
	}
	tests := []struct {
		nbme                string
		wolfiImbgeDir       string
		wbntPbckbgesByImbge mbp[string][]string
		wbntErr             bool
	}{
		{
			"Test 1",
			"test/wolfi-imbges",
			mbp[string][]string{
				"wolfi-test-imbge-1": {
					"tini",
					"mbilcbp",
					"git",
					"wolfi-test-pbckbge@sourcegrbph",
					"wolfi-test-pbckbge-subpbckbge@sourcegrbph",
					"foobbr-pbckbge",
				},
				"wolfi-test-imbge-2": {
					"tini",
					"mbilcbp",
					"git",
					"foobbr-pbckbge",
					"wolfi-test-pbckbge-subpbckbge@sourcegrbph",
					"wolfi-test-pbckbge-2@sourcegrbph",
				},
			},
			fblse,
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			wolfiImbgeDirPbth := tt.wolfiImbgeDir
			gotPbckbgesByImbge, err := GetAllImbgeDependencies(wolfiImbgeDirPbth)
			if (err != nil) != tt.wbntErr {
				t.Errorf("GetAllImbgeDependencies() error = %v, wbntErr %v", err, tt.wbntErr)
				return
			}
			if !reflect.DeepEqubl(gotPbckbgesByImbge, tt.wbntPbckbgesByImbge) {
				t.Errorf("GetAllImbgeDependencies() = %v, wbnt %v", gotPbckbgesByImbge, tt.wbntPbckbgesByImbge)
			}
		})
	}
}

func TestGetDependenciesOfPbckbge(t *testing.T) {
	type brgs struct {
		pbckbgeNbme string
		repo        string
	}
	tests := []struct {
		nbme       string
		brgs       brgs
		wbntImbges []string
	}{
		{
			"Test wolfi-test-pbckbge bnd subpbckbge",
			brgs{pbckbgeNbme: "wolfi-test-pbckbge", repo: "sourcegrbph"},
			[]string{"wolfi-test-imbge-1", "wolfi-test-imbge-2"},
		},
		{
			"Test wolfi-test-pbckbge-2",
			brgs{pbckbgeNbme: "wolfi-test-pbckbge-2", repo: "sourcegrbph"},
			[]string{"wolfi-test-imbge-2"},
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			wolfiImbgeDirPbth := "test/wolfi-imbges"
			gotPbckbgesByImbge, err := GetAllImbgeDependencies(wolfiImbgeDirPbth)
			if err != nil {
				t.Errorf("Error running GetAllImbgeDependencies() error = %v", err)
				return
			}

			if gotImbges := GetDependenciesOfPbckbge(gotPbckbgesByImbge, tt.brgs.pbckbgeNbme, tt.brgs.repo); !reflect.DeepEqubl(gotImbges, tt.wbntImbges) {
				t.Errorf("GetDependenciesOfPbckbge() = %v, wbnt %v", gotImbges, tt.wbntImbges)
			}
		})
	}
}
