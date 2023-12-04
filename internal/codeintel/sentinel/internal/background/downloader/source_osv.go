package downloader

import (
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/shared"

	gocvss20 "github.com/pandatix/go-cvss/20"
	gocvss30 "github.com/pandatix/go-cvss/30"
	gocvss31 "github.com/pandatix/go-cvss/31"
)

// OSV represents the Open Source Vulnerability format.
// See https://ossf.github.io/osv-schema/
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
	Affected   []OSVAffected `json:"affected"`
	References []struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"references"`
	Credits []struct {
		Name    string   `json:"name"`
		Contact []string `json:"contact"`
	} `json:"credits"`
	DatabaseSpecific interface{} `json:"database_specific"` // Provider-specific data, parsed by topLevelHandler
}

// OSVAffected describes packages which are affected by an OSV vulnerability
type OSVAffected struct {
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

	Versions          []string    `json:"versions"`
	EcosystemSpecific interface{} `json:"ecosystem_specific"` // Provider-specific data, parsed by affectedHandler
	DatabaseSpecific  interface{} `json:"database_specific"`  // Provider-specific data, parsed by affectedHandler
}

// DataSourceHandler allows vulnerability database to provide handlers for parsing database-specific data structures.
// Custom data structures can be provided at various locations in OSV, and are named DatabaseSpecific or EcosystemSpecific.
type DataSourceHandler interface {
	topLevelHandler(OSV, *shared.Vulnerability) error           // Handle provider-specific data at the top level of the OSV struct
	affectedHandler(OSVAffected, *shared.AffectedPackage) error // Handle provider-specific data at the OSV.Affected level
}

// osvToVuln converts an OSV-formatted vulnerability to Sourcegraph's internal Vulnerability format
func (parser *CVEParser) osvToVuln(o OSV, dataSourceHandler DataSourceHandler) (vuln shared.Vulnerability, err error) {
	// Core sections:
	//	- /General details
	//  - Severity - TODO, need to loop over
	//	- /Affected
	//  - /References
	//  - Credits
	//  - /Database_specific

	v := shared.Vulnerability{
		SourceID:    o.ID,
		Summary:     o.Summary,
		Details:     o.Details,
		PublishedAt: o.Published,
		ModifiedAt:  &o.Modified,
		WithdrawnAt: &o.Withdrawn,
		Related:     o.Related,
		Aliases:     o.Aliases,
	}

	for _, reference := range o.References {
		v.URLs = append(v.URLs, reference.URL)
	}

	// Parse custom data with a provider-specific handler
	if err := dataSourceHandler.topLevelHandler(o, &v); err != nil {
		return v, err
	}

	if len(o.Severity) > 1 {
		parser.logger.Warn(
			"unexpected number of severity values (>1)",
			log.String("type", "dataWarning"),
			log.String("sourceID", v.SourceID),
			log.String("actualCount", fmt.Sprint(len(o.Severity))),
		)
	}
	for _, severity := range o.Severity {
		v.CVSSVector = severity.Score

		v.CVSSScore, v.Severity, err = parseCVSS(v.CVSSVector)
		if err != nil {
			parser.logger.Warn(
				"could not parse CVSS vector",
				log.String("type", "dataWarning"),
				log.String("sourceID", v.SourceID),
				log.String("cvssVector", v.CVSSVector),
				log.String("err", err.Error()),
			)
		}
	}

	var pas []shared.AffectedPackage
	for _, affected := range o.Affected {
		var ap shared.AffectedPackage

		ap.PackageName = affected.Package.Name
		ap.Language = affected.Package.Ecosystem

		// Parse custom data with a provider-specific handler
		if err := dataSourceHandler.affectedHandler(affected, &ap); err != nil {
			return v, err
		}

		if len(affected.Ranges) > 1 {
			parser.logger.Warn(
				"unexpected number of affected.Ranges (>1)",
				log.String("type", "dataWarning"),
				log.String("sourceID", v.SourceID),
				log.String("actualNumRanges", fmt.Sprint(len(affected.Ranges))),
			)
		}

		// In all observed cases a single range is used, so keep it simple
		for _, affectedRange := range affected.Ranges {
			// Implement dataSourceHandler.affectedRangeHandler here if needed

			for _, event := range affectedRange.Events {
				if event.Introduced != "" {
					ap.VersionConstraint = append(ap.VersionConstraint, ">="+event.Introduced)
				}
				if event.Fixed != "" {
					ap.VersionConstraint = append(ap.VersionConstraint, "<"+event.Fixed)
					ap.Fixed = true
					fixed := event.Fixed
					ap.FixedIn = &fixed
				}
				if event.LastAffected != "" {
					ap.VersionConstraint = append(ap.VersionConstraint, "<="+event.LastAffected)
				}
				if event.Limit != "" {
					ap.VersionConstraint = append(ap.VersionConstraint, "<="+event.Limit)
				}
			}
		}

		if len(affected.Ranges) == 0 && len(affected.Versions) > 0 {
			// A version indicates a precise affected version, so it doesn't make sense to have >1
			if len(affected.Versions) > 1 {
				parser.logger.Warn(
					"unexpected number of affected versions (>1)",
					log.String("type", "dataWarning"),
					log.String("sourceID", v.SourceID),
					log.String("actual", v.CVSSVector),
					log.String("err", err.Error()),
				)
			}
			ap.VersionConstraint = append(ap.VersionConstraint, "="+affected.Versions[0])
		}

		pas = append(pas, ap)
	}

	v.AffectedPackages = pas

	return v, nil
}

func parseCVSS(cvssVector string) (score string, severity string, err error) {
	// Some data sources include trailing slashes
	cleanCvssVector := strings.TrimRight(cvssVector, "/")

	var baseScore float64
	switch {
	case strings.HasPrefix(cvssVector, "CVSS:3.0"):
		cvss, err := gocvss30.ParseVector(cleanCvssVector)
		if err != nil {
			return "", "", err
		}
		baseScore = cvss.BaseScore()

	case strings.HasPrefix(cvssVector, "CVSS:3.1"):
		cvss, err := gocvss31.ParseVector(cleanCvssVector)
		if err != nil {
			return "", "", err
		}
		baseScore = cvss.BaseScore()

	// CVSS v2 does not have prefix, falls into this condition.
	default:
		cvss, err := gocvss20.ParseVector(cleanCvssVector)
		if err != nil {
			return "", "", err
		}
		baseScore = cvss.BaseScore()
	}

	// Implementation of rating is the same across all CVSS versions.
	// Notice CVSS v2.0 does not have a "rating" in its specification,
	// but has been used when CVSS v3 was published.
	severity, err = gocvss31.Rating(baseScore)
	if err != nil {
		return "", "", err
	}

	score = fmt.Sprintf("%.1f", baseScore)
	return score, severity, nil
}
