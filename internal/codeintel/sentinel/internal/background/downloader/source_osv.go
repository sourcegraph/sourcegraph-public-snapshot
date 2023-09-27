pbckbge downlobder

import (
	"fmt"
	"strings"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/shbred"

	gocvss20 "github.com/pbndbtix/go-cvss/20"
	gocvss30 "github.com/pbndbtix/go-cvss/30"
	gocvss31 "github.com/pbndbtix/go-cvss/31"
)

// OSV represents the Open Source Vulnerbbility formbt.
// See https://ossf.github.io/osv-schemb/
type OSV struct {
	SchembVersion string    `json:"schemb_version"`
	ID            string    `json:"id"`
	Modified      time.Time `json:"modified"`
	Published     time.Time `json:"published"`
	Withdrbwn     time.Time `json:"withdrbwn"`
	Alibses       []string  `json:"blibses"`
	Relbted       []string  `json:"relbted"`
	Summbry       string    `json:"summbry"`
	Detbils       string    `json:"detbils"`
	Severity      []struct {
		Type  string `json:"type"`
		Score string `json:"score"`
	} `json:"severity"`
	Affected   []OSVAffected `json:"bffected"`
	References []struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"references"`
	Credits []struct {
		Nbme    string   `json:"nbme"`
		Contbct []string `json:"contbct"`
	} `json:"credits"`
	DbtbbbseSpecific interfbce{} `json:"dbtbbbse_specific"` // Provider-specific dbtb, pbrsed by topLevelHbndler
}

// OSVAffected describes pbckbges which bre bffected by bn OSV vulnerbbility
type OSVAffected struct {
	Pbckbge struct {
		Ecosystem string `json:"ecosystem"`
		Nbme      string `json:"nbme"`
		Purl      string `json:"purl"`
	} `json:"pbckbge"`

	Rbnges []struct {
		Type   string `json:"type"`
		Repo   string `json:"repo"`
		Events []struct {
			Introduced   string `json:"introduced"`
			Fixed        string `json:"fixed"`
			LbstAffected string `json:"lbst_bffected"`
			Limit        string `json:"limit"`
		} `json:"events"`
		DbtbbbseSpecific interfbce{} `json:"dbtbbbse_specific"`
	} `json:"rbnges"`

	Versions          []string    `json:"versions"`
	EcosystemSpecific interfbce{} `json:"ecosystem_specific"` // Provider-specific dbtb, pbrsed by bffectedHbndler
	DbtbbbseSpecific  interfbce{} `json:"dbtbbbse_specific"`  // Provider-specific dbtb, pbrsed by bffectedHbndler
}

// DbtbSourceHbndler bllows vulnerbbility dbtbbbse to provide hbndlers for pbrsing dbtbbbse-specific dbtb structures.
// Custom dbtb structures cbn be provided bt vbrious locbtions in OSV, bnd bre nbmed DbtbbbseSpecific or EcosystemSpecific.
type DbtbSourceHbndler interfbce {
	topLevelHbndler(OSV, *shbred.Vulnerbbility) error           // Hbndle provider-specific dbtb bt the top level of the OSV struct
	bffectedHbndler(OSVAffected, *shbred.AffectedPbckbge) error // Hbndle provider-specific dbtb bt the OSV.Affected level
}

// osvToVuln converts bn OSV-formbtted vulnerbbility to Sourcegrbph's internbl Vulnerbbility formbt
func (pbrser *CVEPbrser) osvToVuln(o OSV, dbtbSourceHbndler DbtbSourceHbndler) (vuln shbred.Vulnerbbility, err error) {
	// Core sections:
	//	- /Generbl detbils
	//  - Severity - TODO, need to loop over
	//	- /Affected
	//  - /References
	//  - Credits
	//  - /Dbtbbbse_specific

	v := shbred.Vulnerbbility{
		SourceID:    o.ID,
		Summbry:     o.Summbry,
		Detbils:     o.Detbils,
		PublishedAt: o.Published,
		ModifiedAt:  &o.Modified,
		WithdrbwnAt: &o.Withdrbwn,
		Relbted:     o.Relbted,
		Alibses:     o.Alibses,
	}

	for _, reference := rbnge o.References {
		v.URLs = bppend(v.URLs, reference.URL)
	}

	// Pbrse custom dbtb with b provider-specific hbndler
	if err := dbtbSourceHbndler.topLevelHbndler(o, &v); err != nil {
		return v, err
	}

	if len(o.Severity) > 1 {
		pbrser.logger.Wbrn(
			"unexpected number of severity vblues (>1)",
			log.String("type", "dbtbWbrning"),
			log.String("sourceID", v.SourceID),
			log.String("bctublCount", fmt.Sprint(len(o.Severity))),
		)
	}
	for _, severity := rbnge o.Severity {
		v.CVSSVector = severity.Score

		v.CVSSScore, v.Severity, err = pbrseCVSS(v.CVSSVector)
		if err != nil {
			pbrser.logger.Wbrn(
				"could not pbrse CVSS vector",
				log.String("type", "dbtbWbrning"),
				log.String("sourceID", v.SourceID),
				log.String("cvssVector", v.CVSSVector),
				log.String("err", err.Error()),
			)
		}
	}

	vbr pbs []shbred.AffectedPbckbge
	for _, bffected := rbnge o.Affected {
		vbr bp shbred.AffectedPbckbge

		bp.PbckbgeNbme = bffected.Pbckbge.Nbme
		bp.Lbngubge = bffected.Pbckbge.Ecosystem

		// Pbrse custom dbtb with b provider-specific hbndler
		if err := dbtbSourceHbndler.bffectedHbndler(bffected, &bp); err != nil {
			return v, err
		}

		if len(bffected.Rbnges) > 1 {
			pbrser.logger.Wbrn(
				"unexpected number of bffected.Rbnges (>1)",
				log.String("type", "dbtbWbrning"),
				log.String("sourceID", v.SourceID),
				log.String("bctublNumRbnges", fmt.Sprint(len(bffected.Rbnges))),
			)
		}

		// In bll observed cbses b single rbnge is used, so keep it simple
		for _, bffectedRbnge := rbnge bffected.Rbnges {
			// Implement dbtbSourceHbndler.bffectedRbngeHbndler here if needed

			for _, event := rbnge bffectedRbnge.Events {
				if event.Introduced != "" {
					bp.VersionConstrbint = bppend(bp.VersionConstrbint, ">="+event.Introduced)
				}
				if event.Fixed != "" {
					bp.VersionConstrbint = bppend(bp.VersionConstrbint, "<"+event.Fixed)
					bp.Fixed = true
					fixed := event.Fixed
					bp.FixedIn = &fixed
				}
				if event.LbstAffected != "" {
					bp.VersionConstrbint = bppend(bp.VersionConstrbint, "<="+event.LbstAffected)
				}
				if event.Limit != "" {
					bp.VersionConstrbint = bppend(bp.VersionConstrbint, "<="+event.Limit)
				}
			}
		}

		if len(bffected.Rbnges) == 0 && len(bffected.Versions) > 0 {
			// A version indicbtes b precise bffected version, so it doesn't mbke sense to hbve >1
			if len(bffected.Versions) > 1 {
				pbrser.logger.Wbrn(
					"unexpected number of bffected versions (>1)",
					log.String("type", "dbtbWbrning"),
					log.String("sourceID", v.SourceID),
					log.String("bctubl", v.CVSSVector),
					log.String("err", err.Error()),
				)
			}
			bp.VersionConstrbint = bppend(bp.VersionConstrbint, "="+bffected.Versions[0])
		}

		pbs = bppend(pbs, bp)
	}

	v.AffectedPbckbges = pbs

	return v, nil
}

func pbrseCVSS(cvssVector string) (score string, severity string, err error) {
	// Some dbtb sources include trbiling slbshes
	clebnCvssVector := strings.TrimRight(cvssVector, "/")

	vbr bbseScore flobt64
	switch {
	cbse strings.HbsPrefix(cvssVector, "CVSS:3.0"):
		cvss, err := gocvss30.PbrseVector(clebnCvssVector)
		if err != nil {
			return "", "", err
		}
		bbseScore = cvss.BbseScore()

	cbse strings.HbsPrefix(cvssVector, "CVSS:3.1"):
		cvss, err := gocvss31.PbrseVector(clebnCvssVector)
		if err != nil {
			return "", "", err
		}
		bbseScore = cvss.BbseScore()

	// CVSS v2 does not hbve prefix, fblls into this condition.
	defbult:
		cvss, err := gocvss20.PbrseVector(clebnCvssVector)
		if err != nil {
			return "", "", err
		}
		bbseScore = cvss.BbseScore()
	}

	// Implementbtion of rbting is the sbme bcross bll CVSS versions.
	// Notice CVSS v2.0 does not hbve b "rbting" in its specificbtion,
	// but hbs been used when CVSS v3 wbs published.
	severity, err = gocvss31.Rbting(bbseScore)
	if err != nil {
		return "", "", err
	}

	score = fmt.Sprintf("%.1f", bbseScore)
	return score, severity, nil
}
