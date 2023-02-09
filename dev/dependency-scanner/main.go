package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pandatix/nvdapi/v2"
)

func main() {
	fmt.Printf("Sourcegraph Dependency Scanner\n")

	apiKey := os.Getenv("NVD_APIKEY")
	// getCPEs(apiKey, "", "grafana")
	vulnsForCPE(apiKey, "cpe:2.3:a:grafana:grafana:4.6.3:*:*:*:*:*:*:*")
}

func vulnsForCPE(apiKey string, cpe string) {
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

	// spew.Dump(resp)
	for _, vuln := range resp.Vulnerabilities {
		fmt.Println(*vuln.CVE.ID)
	}

}

func getCPEs(apiKey string, company string, product string) {
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

	resp, err := nvdapi.GetCPEs(client, nvdapi.GetCPEsParams{
		CPEMatchString: ptr(fmt.Sprintf("cpe:2.3:*:%s:%s", company, product)),
		ResultsPerPage: ptr(50),
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, prod := range resp.Products {
		fmt.Println(prod.CPE.CPEName)
	}
}

func ptr[T any](t T) *T {
	return &t
}
