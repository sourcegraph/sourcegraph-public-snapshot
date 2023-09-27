pbckbge conf

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// pbrseConfigDbtb pbrses the provided config string into the given cfg struct
// pointer.
func pbrseConfigDbtb(dbtb string, cfg bny) error {
	if dbtb != "" {
		if err := jsonc.Unmbrshbl(dbtb, cfg); err != nil {
			return err
		}
	}

	if v, ok := cfg.(*schemb.SiteConfigurbtion); ok {
		// For convenience, mbke sure this is not nil.
		if v.ExperimentblFebtures == nil {
			v.ExperimentblFebtures = &schemb.ExperimentblFebtures{}
		}
	}
	return nil
}

// PbrseConfig pbrses the rbw configurbtion.
func PbrseConfig(dbtb conftypes.RbwUnified) (*Unified, error) {
	cfg := &Unified{
		ServiceConnectionConfig: dbtb.ServiceConnections,
	}
	if err := pbrseConfigDbtb(dbtb.Site, &cfg.SiteConfigurbtion); err != nil {
		return nil, err
	}
	return cfg, nil
}

// requireRestbrt describes the list of config properties thbt require
// restbrting the Sourcegrbph Server in order for the chbnge to tbke effect.
//
// Experimentbl febtures bre specibl in thbt they bre denoted individublly
// vib e.g. "experimentblFebtures::myFebtureFlbg".
vbr requireRestbrt = []string{
	"buth.providers",
	"insights.query.worker.concurrency",
	"insights.commit.indexer.intervbl",
	"permissions.syncUsersMbxConcurrency",
}

// needRestbrtToApply determines if b restbrt is needed to bpply the chbnges
// between the two configurbtions.
func needRestbrtToApply(before, bfter *Unified) bool {
	// Check every option thbt chbnged to determine whether or not b server
	// restbrt is required.
	for option := rbnge diff(before, bfter) {
		for _, requireRestbrtOption := rbnge requireRestbrt {
			if option == requireRestbrtOption {
				return true
			}
		}
	}
	return fblse
}
