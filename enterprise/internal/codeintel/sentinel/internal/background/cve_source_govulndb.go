package background

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/shared"
)

func ReadGoVulnDb(ctx context.Context) ([]shared.Vulnerability, error) {
	// TODO: Fetch database

	// Open test directory of json files
	path := "./go-vulndb/"
	fileList, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var vulns []shared.Vulnerability
	for _, file := range fileList {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		fullPath := filepath.Join(path, file.Name())

		fmt.Printf("Walking %s\n", file.Name())

		r, err := os.Open(fullPath)
		if err != nil {
			return nil, err
		}
		defer r.Close()

		var osvVuln OSV

		if err := json.NewDecoder(r).Decode(&osvVuln); err != nil {
			return nil, err
		}

		v, err := osvToVuln(osvVuln)
		if err != nil {
			return nil, err
		}

		out, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return nil, err
		}
		fmt.Printf("%s\n\n", string(out))
	}

	return vulns, nil
}

// Open Source Vulnerability format
// https://ossf.github.io/osv-schema/
type OSV struct {
	SchemaVersion string    `json:"schema_version"`
	ID            string    `json:"id"`
	Modified      time.Time `json:"modified"`
	Published     time.Time `json:"published"`
	Withdrawn     time.Time `json:"withdrawn"`
	Aliases       []string  `json:"aliases"`
	Related       []string  `json:"related"`
	Summary       string    `json:"summary"`
	Details       string    `json:"details"`
	Severity      []struct {
		Type  string `json:"type"`
		Score string `json:"score"`
	} `json:"severity"`
	Affected []struct {
		Package struct {
			Ecosystem string `json:"ecosystem"`
			Name      string `json:"name"`
			Purl      string `json:"purl"`
		} `json:"package"`
		Ranges []struct {
			Type   string `json:"type"`
			Repo   string `json:"repo"`
			Events []struct {
				Introduced   string `json:"introduced"`
				Fixed        string `json:"fixed"`
				LastAffected string `json:"last_affected"`
				Limit        string `json:"limit"`
			} `json:"events"`
			DatabaseSpecific interface{} `json:"database_specific"`
		} `json:"ranges"`
		Versions []string `json:"versions"`
		// EcosystemSpecific interface{} `json:"ecosystem_specific"`
		EcosystemSpecific GoVulnDBAffectedEcosystemSpecific `json:"ecosystem_specific"`
		// DatabaseSpecific  interface{}       `json:"database_specific"`
		DatabaseSpecific map[string]string `json:"database_specific"` // TODO: Currently hardcoding GoVulndb format
	} `json:"affected"`
	References []struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"references"`
	Credits []struct {
		Name    string   `json:"name"`
		Contact []string `json:"contact"`
	} `json:"credits"`
	DatabaseSpecific interface{} `json:"database_specific"`
}

type GoVulnDBAffectedcDatabaseSpecific map[string]string

type GoVulnDBAffectedEcosystemSpecific struct {
	Imports []struct {
		Path    string   `json:"path"`
		Symbols []string `json:"symbols"`
	} `json:"imports"`
}

func osvToVuln(o OSV) (vuln shared.Vulnerability, err error) {
	v := shared.Vulnerability{
		SourceID:  o.ID,
		Summary:   o.Summary,
		Details:   o.Details,
		Published: o.Published,
		Modified:  o.Modified,
		Withdrawn: o.Withdrawn,
		// CWEs:                   o.DatabaseSpecific.CWEIDs,
		Related: o.Related,
		Aliases: o.Aliases,
	}

	for _, reference := range o.References {
		v.URLs = append(v.URLs, reference.URL)
	}

	for _, affected := range o.Affected {
		var pa shared.AffectedPackage

		pa.PackageName = affected.Package.Name
		pa.Language = affected.Package.Ecosystem
		pa.Namespace = "govulndb" // TODO:

		v.DataSource = affected.DatabaseSpecific["url"]

		for _, r := range affected.Ranges {
			_ = r
		}

		for _, i := range affected.EcosystemSpecific.Imports {
			pa.AffectedSymbols = append(pa.AffectedSymbols, shared.AffectedSymbol{
				Path:    i.Path,
				Symbols: i.Symbols,
			})
		}

		v.AffectedPackages = append(v.AffectedPackages, pa)
	}

	return v, nil
}
