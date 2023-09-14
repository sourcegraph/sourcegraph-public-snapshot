package downloader

// Parse vulnerabilities from the golang/VulnDB (Govulndb) database.
// Govulndb uses the Open Source Vulnerability (OSV) format, with some custom extensions.

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mitchellh/mapstructure"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/shared"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const govulndbAdvisoryDatabaseURL = "https://github.com/golang/vuln/archive/refs/heads/master.zip"

// ReadGoVulnDb fetches a copy of the Go Vulnerability Database and converts it to the internal Vulnerability format
func (parser *CVEParser) ReadGoVulnDb(ctx context.Context, useLocalCache bool) (vulns []shared.Vulnerability, err error) {
	if useLocalCache {
		zipReader, err := os.Open("vulndb-govulndb.zip")
		if err != nil {
			return nil, errors.New("unable to open zip file")
		}

		return parser.ParseGovulndbAdvisoryDB(zipReader)
	}

	resp, err := http.Get(govulndbAdvisoryDatabaseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Newf("unexpected status code %d", resp.StatusCode)
	}

	return parser.ParseGitHubAdvisoryDB(resp.Body)
}

func (parser *CVEParser) ParseGovulndbAdvisoryDB(govulndbReader io.Reader) (vulns []shared.Vulnerability, err error) {
	content, err := io.ReadAll(govulndbReader)
	if err != nil {
		return nil, err
	}

	zr, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return nil, err
	}

	for _, f := range zr.File {
		if filepath.Dir(f.Name) != "vulndb-master/data/osv" {
			continue
		}
		if filepath.Ext(f.Name) != ".json" {
			continue
		}

		r, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer r.Close()

		var osvVuln OSV
		if err := json.NewDecoder(r).Decode(&osvVuln); err != nil {
			return nil, err
		}

		// Convert OSV to Vulnerability using Govulndb handler
		var g Govulndb
		convertedVuln, err := parser.osvToVuln(osvVuln, g)
		if err != nil {
			return nil, err
		}

		vulns = append(vulns, convertedVuln)
	}

	return vulns, nil
}

//
// Govulndb-specific structs and handlers
//

// GovulndbAffectedEcosystemSpecific represents the custom data format used by Govulndb for OSV.Affected.EcosystemSpecific
type GovulndbAffectedEcosystemSpecific struct {
	Imports []struct {
		Path    string   `mapstructure:"path" json:"path"`
		Goos    []string `mapstructure:"goos" json:"goos"`
		Symbols []string `mapstructure:"symbols" json:"symbols"`
	} `mapstructure:"imports" json:"imports"`
}

// GovulndbAffectedDatabaseSpecific represents the custom data format used by Govulndb for OSV.Affected.DatabaseSpecific
type GovulndbAffectedDatabaseSpecific struct {
	URL string `json:"url"`
}

type Govulndb int64

func (g Govulndb) topLevelHandler(o OSV, v *shared.Vulnerability) error {
	v.DataSource = "https://pkg.go.dev/vuln/" + o.ID

	// Govulndb doesn't provide any top-level database_specific data
	return nil
}

func (g Govulndb) affectedHandler(a OSVAffected, affectedPackage *shared.AffectedPackage) error {
	affectedPackage.Namespace = "govulndb"

	// Attempt to decode the JSON from an interface{} to GovulnDBAffectedEcosystemSpecific
	var es GovulndbAffectedEcosystemSpecific
	if err := mapstructure.Decode(a.EcosystemSpecific, &es); err != nil {
		return errors.Wrap(err, "cannot map DatabaseSpecific to GovulndbAffectedEcosystemSpecific")
	}

	for _, i := range es.Imports {
		affectedPackage.AffectedSymbols = append(affectedPackage.AffectedSymbols, shared.AffectedSymbol{
			Path:    i.Path,
			Symbols: i.Symbols,
		})
	}

	return nil
}
