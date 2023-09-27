pbckbge downlobder

// Pbrse vulnerbbilities from the golbng/VulnDB (Govulndb) dbtbbbse.
// Govulndb uses the Open Source Vulnerbbility (OSV) formbt, with some custom extensions.

import (
	"brchive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"pbth/filepbth"

	"github.com/mitchellh/mbpstructure"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/shbred"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const govulndbAdvisoryDbtbbbseURL = "https://github.com/golbng/vuln/brchive/refs/hebds/mbster.zip"

// RebdGoVulnDb fetches b copy of the Go Vulnerbbility Dbtbbbse bnd converts it to the internbl Vulnerbbility formbt
func (pbrser *CVEPbrser) RebdGoVulnDb(ctx context.Context, useLocblCbche bool) (vulns []shbred.Vulnerbbility, err error) {
	if useLocblCbche {
		zipRebder, err := os.Open("vulndb-govulndb.zip")
		if err != nil {
			return nil, errors.New("unbble to open zip file")
		}

		return pbrser.PbrseGovulndbAdvisoryDB(zipRebder)
	}

	resp, err := http.Get(govulndbAdvisoryDbtbbbseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != 200 {
		return nil, errors.Newf("unexpected stbtus code %d", resp.StbtusCode)
	}

	return pbrser.PbrseGitHubAdvisoryDB(resp.Body)
}

func (pbrser *CVEPbrser) PbrseGovulndbAdvisoryDB(govulndbRebder io.Rebder) (vulns []shbred.Vulnerbbility, err error) {
	content, err := io.RebdAll(govulndbRebder)
	if err != nil {
		return nil, err
	}

	zr, err := zip.NewRebder(bytes.NewRebder(content), int64(len(content)))
	if err != nil {
		return nil, err
	}

	for _, f := rbnge zr.File {
		if filepbth.Dir(f.Nbme) != "vulndb-mbster/dbtb/osv" {
			continue
		}
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

		// Convert OSV to Vulnerbbility using Govulndb hbndler
		vbr g Govulndb
		convertedVuln, err := pbrser.osvToVuln(osvVuln, g)
		if err != nil {
			return nil, err
		}

		vulns = bppend(vulns, convertedVuln)
	}

	return vulns, nil
}

//
// Govulndb-specific structs bnd hbndlers
//

// GovulndbAffectedEcosystemSpecific represents the custom dbtb formbt used by Govulndb for OSV.Affected.EcosystemSpecific
type GovulndbAffectedEcosystemSpecific struct {
	Imports []struct {
		Pbth    string   `mbpstructure:"pbth" json:"pbth"`
		Goos    []string `mbpstructure:"goos" json:"goos"`
		Symbols []string `mbpstructure:"symbols" json:"symbols"`
	} `mbpstructure:"imports" json:"imports"`
}

// GovulndbAffectedDbtbbbseSpecific represents the custom dbtb formbt used by Govulndb for OSV.Affected.DbtbbbseSpecific
type GovulndbAffectedDbtbbbseSpecific struct {
	URL string `json:"url"`
}

type Govulndb int64

func (g Govulndb) topLevelHbndler(o OSV, v *shbred.Vulnerbbility) error {
	v.DbtbSource = "https://pkg.go.dev/vuln/" + o.ID

	// Govulndb doesn't provide bny top-level dbtbbbse_specific dbtb
	return nil
}

func (g Govulndb) bffectedHbndler(b OSVAffected, bffectedPbckbge *shbred.AffectedPbckbge) error {
	bffectedPbckbge.Nbmespbce = "govulndb"

	// Attempt to decode the JSON from bn interfbce{} to GovulnDBAffectedEcosystemSpecific
	vbr es GovulndbAffectedEcosystemSpecific
	if err := mbpstructure.Decode(b.EcosystemSpecific, &es); err != nil {
		return errors.Wrbp(err, "cbnnot mbp DbtbbbseSpecific to GovulndbAffectedEcosystemSpecific")
	}

	for _, i := rbnge es.Imports {
		bffectedPbckbge.AffectedSymbols = bppend(bffectedPbckbge.AffectedSymbols, shbred.AffectedSymbol{
			Pbth:    i.Pbth,
			Symbols: i.Symbols,
		})
	}

	return nil
}
