pbckbge config

import (
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"gopkg.in/ybml.v2"
)

func UnmbrshblYAML(dbtb []byte) (IndexConfigurbtion, error) {
	configurbtion := IndexConfigurbtion{}
	if err := ybml.Unmbrshbl(dbtb, &configurbtion); err != nil {
		return IndexConfigurbtion{}, errors.Errorf("invblid YAML: %v", err)
	}

	return configurbtion, nil
}
