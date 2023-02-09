package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

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

type Vulnerability struct {
	CVE                  string `json:"cve"`
	Description          string `json:"description"`
	Dependency           string `json:"dependency"`
	PackageManager       string `json:"package_manager"`
	PublishedDate        string `json:"published_date"`
	LastUpdate           string `json:"last_update"`
	SourceFile           string `json:"source_file"`
	SourceFileLineNumber int    `json:"source_file_line_number"`
	AffectedVersion      string `json:"affected_version"`
	CurrentVersion       string `json:"current_version"`
	SeverityScore        string `json:"severity_score"`
	SeverityString       string `json:"severity_string"`
}

func main() {
	fmt.Printf("Sourcegraph Dependency Scanner\n")

	apiKey := os.Getenv("NVD_APIKEY")

	deps, err := parseDependencies("dependencies-js.json")
	if err != nil {
		log.Fatal(err)
	}

	for _, dep := range deps {
		if dep.Manager != "npm" {
			fmt.Printf("Skipping non-npm dependency\n")
		}

		vulns := jsDepToVuln(apiKey, dep.Name, dep.Version)
		printCVEs(vulns)
	}

	// jsDepToVuln(apiKey, "electerm", "1.3.22")
	// jsDepToVuln(apiKey, "lodash", "4.17.20")

	return

	// getCPEs(apiKey, "", "grafana")
	// vulnsForCPE(apiKey, "cpe:2.3:a:grafana:grafana:4.6.3:*:*:*:*:*:*:*")

	// deps, err := parseDependencies("dependencies-simple-3.json")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // spew.Dump(deps)
	// printDeps(deps)
}

func jsDepToVuln(apiKey string, inputProduct string, inputVersion string) []nvdapi.CVEItem {
	fmt.Printf("Looking up %s version %s\n", inputProduct, inputVersion)
	author, product := lookupJSDependencyCPE(apiKey, inputProduct)
	cpeString := fmt.Sprintf("cpe:2.3:a:%s:%s:%s:*:*:*:*:*:*:*", author, product, inputVersion)
	fmt.Printf("CPE string for %s version %s is %s\n", inputProduct, inputVersion, cpeString)

	vulns := vulnsForCPE(apiKey, cpeString)

	return vulns
}

// Guess the CPE for a JS dependency
// e.g. Electerm -> electerm_project : electerm
func lookupJSDependencyCPE(apiKey string, name string) (author string, product string) {
	fmt.Printf("Looking up CPE for JS dependency %s\n", name)

	productCPEs := getCPEs(apiKey, "", name, "")
	// printCPEProducts(productCPEs)

	// Take the first CPE and extract the author field - this is a quick and reasonable guess
	author, product, err := parseCPE(productCPEs[0].CPE.CPEName)
	if err != nil {
		log.Fatal(err)
	}

	return author, product
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

func printDeps(deps Dependencies) {
	for _, dep := range deps {
		fmt.Printf("* %s version %s (%s)\n\n", dep.Name, dep.Version, dep.Manager)
	}
}

func printCPEProducts(products ProductCPEs) {
	for _, prod := range products {
		fmt.Println(prod.CPE.CPEName)
	}
}

func printCVEs(vulns []nvdapi.CVEItem) {
	// spew.Dump(resp)
	for _, vuln := range vulns {
		fmt.Println(*vuln.CVE.ID)
	}
}

func parseDependencies(dependencyFilePath string) (deps []Dependency, err error) {
	jsonFile, err := os.Open(dependencyFilePath)
	if err != nil {
		log.Fatal("Unable to open dependency file")
	}
	defer jsonFile.Close()

	jsonBytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatal("Unable to read dependency file")
	}

	// var dependencyFile DependencyFile
	// var dep Dependency
	var depFile DependencyFile
	json.Unmarshal(jsonBytes, &depFile)

	deps = depFile["data"]["dependencies"]

	return deps, nil
}

func vulnsForCPE(apiKey string, cpe string) []nvdapi.CVEItem {
	client, err := nvdapi.NewNVDClient(&http.Client{}, apiKey)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := nvdapi.GetCVEs(client, nvdapi.GetCVEsParams{
		CPEName: ptr(cpe),
	})
	if err != nil {
		log.Fatal(err)
	}

	return resp.Vulnerabilities
}

func getCPEs(apiKey string, company string, product string, version string) (cpes ProductCPEs) {
	client, err := nvdapi.NewNVDClient(&http.Client{}, apiKey)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}

	return resp.Products
}

func ptr[T any](t T) *T {
	return &t
}
