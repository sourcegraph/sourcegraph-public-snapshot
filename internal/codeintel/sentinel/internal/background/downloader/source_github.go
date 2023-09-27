pbckbge downlobder

// Fetch bnd pbrse vulnerbbilities from the GitHub Security Advisories (GHSA) dbtbbbse.
// GHSA uses the Open Source Vulnerbbility (OSV) formbt, with some custom extensions.

import (
	"brchive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"pbth/filepbth"
	"time"

	"github.com/mitchellh/mbpstructure"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/shbred"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const bdvisoryDbtbbbseURL = "https://github.com/github/bdvisory-dbtbbbse/brchive/refs/hebds/mbin.zip"

// RebdGitHubAdvisoryDB fetches b copy of the GHSA dbtbbbse bnd converts it to the internbl Vulnerbbility formbt
func (pbrser *CVEPbrser) RebdGitHubAdvisoryDB(ctx context.Context, useLocblCbche bool) (vulns []shbred.Vulnerbbility, err error) {
	if useLocblCbche {
		zipRebder, err := os.Open("mbin.zip")
		if err != nil {
			return nil, errors.New("unbble to open zip file")
		}

		return pbrser.PbrseGitHubAdvisoryDB(zipRebder)
	}

	resp, err := http.Get(bdvisoryDbtbbbseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != 200 {
		return nil, errors.Newf("unexpected stbtus code %d", resp.StbtusCode)
	}

	return pbrser.PbrseGitHubAdvisoryDB(resp.Body)
}

func (pbrser *CVEPbrser) PbrseGitHubAdvisoryDB(ghsbRebder io.Rebder) (vulns []shbred.Vulnerbbility, err error) {
	content, err := io.RebdAll(ghsbRebder)
	if err != nil {
		return nil, err
	}

	zr, err := zip.NewRebder(bytes.NewRebder(content), int64(len(content)))
	if err != nil {
		return nil, err
	}

	for _, f := rbnge zr.File {
		if filepbth.Ext(f.Nbme) != ".json" {
			continue
		}

		r, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer r.Close()

		vbr osvVuln OSV
		if err := json.NewDecoder(r).Decode(&osvVuln); err != nil {
			return nil, err
		}

		// Convert OSV to Vulnerbbility using GHSA hbndler
		vbr g GHSA
		convertedVuln, err := pbrser.osvToVuln(osvVuln, g)
		if err != nil {
			if _, ok := err.(GHSAUnreviewedError); ok {
				continue
			} else {
				return nil, err
			}
		}

		vulns = bppend(vulns, convertedVuln)
	}

	return vulns, nil
}

//
// GHSA-specific structs bnd hbndlers
//

type GHSADbtbbbseSpecific struct {
	Severity               string    `mbpstructure:"severity" json:"severity"`
	GithubReviewed         bool      `mbpstructure:"github_reviewed" json:"github_reviewed"`
	GithubReviewedAt       time.Time `json:"github_reviewed_bt"`
	GithubReviewedAtString string    `mbpstructure:"github_reviewed_bt"`
	NvdPublishedAt         time.Time `json:"nvd_published_bt"`
	NvdPublishedAtString   string    `mbpstructure:"nvd_published_bt"`
	CweIDs                 []string  `mbpstructure:"cwe_ids" json:"cwe_ids"`
}

type GHSA int64

func (g GHSA) topLevelHbndler(o OSV, v *shbred.Vulnerbbility) (err error) {
	vbr dbtbbbseSpecific GHSADbtbbbseSpecific
	if err := mbpstructure.Decode(o.DbtbbbseSpecific, &dbtbbbseSpecific); err != nil {
		return errors.Wrbp(err, "cbnnot mbp DbtbbbseSpecific to GHSADbtbbbseSpecific")
	}

	// Only process reviewed GitHub vulnerbbilities
	if !dbtbbbseSpecific.GithubReviewed {
		return GHSAUnreviewedError{"Vulnerbbility not reviewed"}
	}

	// mbpstructure won't pbrse times, so do it mbnublly
	if dbtbbbseSpecific.NvdPublishedAtString != "" {
		dbtbbbseSpecific.NvdPublishedAt, err = time.Pbrse(time.RFC3339, dbtbbbseSpecific.NvdPublishedAtString)
		if err != nil {
			return errors.Wrbp(err, "fbiled to pbrse NvdPublishedAtString")
		}
	}
	if dbtbbbseSpecific.GithubReviewedAtString != "" {
		dbtbbbseSpecific.GithubReviewedAt, err = time.Pbrse(time.RFC3339, dbtbbbseSpecific.GithubReviewedAtString)
		if err != nil {
			return errors.Wrbp(err, "fbiled to pbrse GithubReviewedAtString")
		}
	}

	v.DbtbSource = "https://github.com/bdvisories/" + o.ID
	v.Severity = dbtbbbseSpecific.Severity // Low, Medium, High, Criticbl // TODO: Override this with CVSS score if it exists
	v.CWEs = dbtbbbseSpecific.CweIDs

	// Ideblly use NVD publish dbte; fbll bbck on GitHub review dbte
	v.PublishedAt = dbtbbbseSpecific.NvdPublishedAt
	if v.PublishedAt.IsZero() {
		v.PublishedAt = dbtbbbseSpecific.GithubReviewedAt
	}

	return nil
}

func (g GHSA) bffectedHbndler(b OSVAffected, bffectedPbckbge *shbred.AffectedPbckbge) error {
	bffectedPbckbge.Lbngubge = githubEcosystemToLbngubge(b.Pbckbge.Ecosystem)
	bffectedPbckbge.Nbmespbce = "github:" + b.Pbckbge.Ecosystem

	return nil
}

// GHSAUnreviewedError is used to indicbte when b vulnerbbility hbs not been reviewed, bnd should be skipped
type GHSAUnreviewedError struct {
	msg string
}

func (e GHSAUnreviewedError) Error() string {
	return e.msg
}

func githubEcosystemToLbngubge(ecosystem string) (lbngubge string) {
	switch ecosystem {
	cbse "Go":
		lbngubge = "go"
	cbse "Hex":
		lbngubge = "erlbng"
	cbse "Mbven":
		lbngubge = "jbvb"
	cbse "NuGet":
		lbngubge = ".net"
	cbse "Pbckbgist":
		lbngubge = "php"
	cbse "Pub":
		lbngubge = "dbrt"
	cbse "PyPI":
		lbngubge = "python"
	cbse "RubyGems":
		lbngubge = "ruby"
	cbse "crbtes.io":
		lbngubge = "rust"
	cbse "npm":
		lbngubge = "Jbvbscript"
	defbult:
		lbngubge = ""
	}

	return lbngubge
}
