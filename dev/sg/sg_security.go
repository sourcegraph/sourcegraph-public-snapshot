package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
)

var securityCommand = &cli.Command{
	Name:        "security",
	Usage:       "interact with Sourcegraph security tools",
	Description: "Learn more about Sourcegraph security: https://about.sourcegraph.com/security",
	Category:    category.Company,
	Subcommands: []*cli.Command{
		{
			Name:        "repoferee-report",
			Usage:       "fetch the latest repoferee report",
			Description: "To learn more about Cloud V2, see https://handbook.sourcegraph.com/departments/cloud/technical-docs/v2.0/",
			Action:      getRepofereeReport,
		},
	},
}

const RepofereeEndpoint = "https://repoferee.sgdev.org"

func hmacSign(key, message string) []byte {
	hasher := hmac.New(sha256.New, []byte(key))
	hasher.Write([]byte(message))
	return hasher.Sum(nil)
}

func getRepofereeReport(c *cli.Context) error {
	authorization := "Bearer 685013726e4bbe8e79f5d6c21902eb66a1e9549fed8e60b0d7efbc7305c4f7f6"
	timestamp := time.Now().Format(time.RFC3339)
	path := "/results"

	hmac := "99c045e02363d6ba1742ea8737546cf64597d937292ba3e416a7cd03d2ffa404"
	urlPath, err := url.JoinPath(RepofereeEndpoint, path)
	if err != nil {
		return err
	}
	signature := hmacSign(hmac, fmt.Sprintf("%s:%s:%s", timestamp, path, authorization))
	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", authorization)
	req.Header.Add("X-Timestamp", timestamp)
	req.Header.Add("X-Signature", hex.EncodeToString(signature))
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	value := struct {
		Results map[string]json.RawMessage `json:"results"`
		LastRun string                     `json:"last_run_at"`
		NextRun string                     `json:"next_run_at"`
	}{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&value); err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent(" ", " ")
	if err := enc.Encode(value); err != nil {
		return err
	}

	return nil
}
