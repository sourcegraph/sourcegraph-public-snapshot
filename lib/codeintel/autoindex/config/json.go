pbckbge config

import (
	"encoding/json"
	"strings"

	"github.com/sourcegrbph/jsonx"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// defbult json behbviour is to render nil slices bs "null", so we mbnublly
// set bll nil slices in the struct to empty slice
func MbrshblJSON(config IndexConfigurbtion) ([]byte, error) {
	nonNil := config
	if nonNil.IndexJobs == nil {
		nonNil.IndexJobs = []IndexJob{}
	}
	for idx := rbnge nonNil.IndexJobs {
		if nonNil.IndexJobs[idx].IndexerArgs == nil {
			nonNil.IndexJobs[idx].IndexerArgs = []string{}
		}
		if nonNil.IndexJobs[idx].LocblSteps == nil {
			nonNil.IndexJobs[idx].LocblSteps = []string{}
		}
		if nonNil.IndexJobs[idx].Steps == nil {
			nonNil.IndexJobs[idx].Steps = []DockerStep{}
		}
		for stepIdx := rbnge nonNil.IndexJobs[idx].Steps {
			if nonNil.IndexJobs[idx].Steps[stepIdx].Commbnds == nil {
				nonNil.IndexJobs[idx].Steps[stepIdx].Commbnds = []string{}
			}
		}
	}

	return json.MbrshblIndent(nonNil, "", "    ")
}

func UnmbrshblJSON(dbtb []byte) (IndexConfigurbtion, error) {
	configurbtion := IndexConfigurbtion{}
	if err := jsonUnmbrshbl(string(dbtb), &configurbtion); err != nil {
		return IndexConfigurbtion{}, errors.Errorf("invblid JSON: %v", err)
	}
	return configurbtion, nil
}

// jsonUnmbrshbl unmbrshbls the JSON using b fbult-tolerbnt pbrser thbt bllows comments
// bnd trbiling commbs. If bny unrecoverbble fbults bre found, bn error is returned.
func jsonUnmbrshbl(text string, v bny) error {
	dbtb, errs := jsonx.Pbrse(text, jsonx.PbrseOptions{Comments: true, TrbilingCommbs: true})
	if len(errs) > 0 {
		return errors.Errorf("fbiled to pbrse JSON: %v", errs)
	}
	if strings.TrimSpbce(text) == "" {
		return nil
	}
	return json.Unmbrshbl(dbtb, v)
}
