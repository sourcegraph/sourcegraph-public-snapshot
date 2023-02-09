package graphql

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/pandatix/nvdapi/v2"
)

// type DependencyFileTop map[string]DependencyFile
type DependencyFile map[string]map[string]Dependencies

type Dependencies []Dependency

type Dependency struct {
	Manager string `json:"manager"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ProductCPEs []nvdapi.CPEProduct
type NVDCVEs []nvdapi.CVEItem

type Vulnerability struct {
	CVE             string `json:"cve"`
	Description     string `json:"description"`
	Dependency      string `json:"dependency"`
	PackageManager  string `json:"packageManager"`
	PublishedDate   string `json:"publishedDate"`
	LastUpdate      string `json:"lastUpdate"`
	AffectedVersion string `json:"affectedVersion"`
	CurrentVersion  string `json:"currentVersion"`
	SeverityScore   string `json:"severityScore"`
	SeverityString  string `json:"severityString"`
}

var apiKey = os.Getenv("NVD_APIKEY")

// convertToVulnerabilities converts from the API response format to the struct used by our frontend
func convertToVulnerabilities(inputProduct string, inputVersion string, cveitems []nvdapi.CVEItem) (vs []Vulnerability) {
	for _, cveitem := range cveitems {
		spew.Dump(cveitem)

		c := cveitem.CVE

		v := Vulnerability{
			CVE:             *c.ID,
			Description:     c.Descriptions[0].Value,
			Dependency:      inputProduct,
			PackageManager:  "npm",
			PublishedDate:   *c.Published,
			LastUpdate:      *c.LastModified,
			AffectedVersion: "unknown",
			CurrentVersion:  inputVersion,
			SeverityScore:   fmt.Sprintf("%.1f", c.Metrics.CVSSMetricV31[0].CVSSData.BaseScore),
			SeverityString:  strings.Title(strings.ToLower(c.Metrics.CVSSMetricV31[0].CVSSData.BaseSeverity)),
		}

		// CVSSData <- CVSSMetricsV31 <- []Metrics

		vs = append(vs, v)
	}

	return vs
}

func jsDepToVuln(apiKey string, inputProduct string, inputVersion string) ([]Vulnerability, error) {
	fmt.Printf("Looking up %s version %s\n", inputProduct, inputVersion)
	author, product, err := lookupJSDependencyCPE(apiKey, inputProduct)
	if err != nil {
		return nil, err
	}
	if author == "" && product == "" {
		return nil, nil
	}
	cpeString := fmt.Sprintf("cpe:2.3:a:%s:%s:%s:*:*:*:*:*:*:*", author, product, inputVersion)
	fmt.Printf("CPE string for %s version %s is %s\n", inputProduct, inputVersion, cpeString)

	vulns, err := vulnsForCPE(apiKey, cpeString)
	if err != nil {
		return nil, err
	}

	vs := convertToVulnerabilities(inputProduct, inputVersion, vulns)
	return vs, nil
}

// Guess the CPE for a JS dependency
// e.g. Electerm -> electerm_project : electerm
func lookupJSDependencyCPE(apiKey string, name string) (author string, product string, _ error) {
	fmt.Printf("Looking up CPE for JS dependency %s\n", name)

	productCPEs, err := getCPEs(apiKey, "", name, "")
	if err != nil {
		return "", "", err
	}
	if len(productCPEs) == 0 {
		return "", "", nil
	}

	// printCPEProducts(productCPEs)

	// Take the first CPE and extract the author field - this is a quick and reasonable guess
	author, product, err = parseCPE(productCPEs[0].CPE.CPEName)
	if err != nil {
		return "", "", err
	}

	return author, product, err
}

func parseCPE(cpe string) (author string, product string, err error) {
	cpeRegex := regexp.MustCompile(`^cpe:2.3:a:([^:]+):([^:]+):`)

	matches := cpeRegex.FindStringSubmatch(cpe)
	if len(matches) == 0 {
		fmt.Printf("CPE regex does not match")
		return "", "", errors.New("CPE regex does not match")
	}

	fmt.Printf("Author: %s, product: %s\n", matches[1], matches[2])

	return matches[1], matches[2], nil
}

func vulnsForCPE(apiKey string, cpe string) ([]nvdapi.CVEItem, error) {
	client, err := nvdapi.NewNVDClient(&http.Client{}, apiKey)
	if err != nil {
		return nil, err
	}

	resp, err := nvdapi.GetCVEs(client, nvdapi.GetCVEsParams{
		CPEName: ptr(cpe),
	})
	if err != nil {
		return nil, err
	}

	return resp.Vulnerabilities, nil
}

func getCPEs(apiKey string, company string, product string, version string) (cpes ProductCPEs, _ error) {
	client, err := nvdapi.NewNVDClient(&http.Client{}, apiKey)
	if err != nil {
		return nil, err
	}

	if company == "" {
		company = "*"
	}
	if product == "" {
		product = "*"
	}
	if version == "" {
		version = "*"
	}

	resp, err := nvdapi.GetCPEs(client, nvdapi.GetCPEsParams{
		CPEMatchString: ptr(fmt.Sprintf("cpe:2.3:*:%s:%s:%s", company, product, version)),
		ResultsPerPage: ptr(50),
	})
	if err != nil {
		return nil, err
	}

	return resp.Products, nil
}

func ptr[T any](t T) *T {
	return &t
}
