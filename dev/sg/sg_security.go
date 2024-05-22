package main

import (
	"context"
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
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var securityCommand = &cli.Command{
	Name:        "security",
	Usage:       "interact with Sourcegraph security tools",
	Description: "Learn more about Sourcegraph security: https://sourcegraph.notion.site/Security-81d50b5ac5474b07bdbadd5359993c80",
	Category:    category.Company,
	Subcommands: []*cli.Command{
		{
			Name:        "repo-report",
			Usage:       "fetch the latest repoferee report",
			Description: "Fetches the latest repoferee report which reports on repositories that conform to set rules. The rules can be found at https://github.com/sourcegraph/repoferee/blob/main/rules.yml",
			Action:      getRepofereeReport,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "output-file",
					DefaultText: "standard out",
					Usage:       "<filename> to write the report to",
					Value:       "",
					Aliases:     []string{"o"},
				},
			},
		},
	},
}

const RepofereeEndpoint = "https://repoferee.sgdev.org"

type repofereeSecret struct {
	AuthToken string
	HMAC      string
}

func hmacSign(key, message string) []byte {
	hasher := hmac.New(sha256.New, []byte(key))
	hasher.Write([]byte(message))
	return hasher.Sum(nil)
}

func getRepofereeSecrets(ctx context.Context) (repofereeSecret, error) {
	var values repofereeSecret
	store, err := secrets.FromContext(ctx)
	if err != nil {
		return values, err
	}

	sec := secrets.ExternalSecret{
		Project: secrets.LocalDevProject,
		Name:    "SG_REPOFEREE_AUTH_TOKEN",
	}
	values.AuthToken, err = store.GetExternal(ctx, sec)
	if err != nil {
		return values, err
	}

	sec = secrets.ExternalSecret{
		Project: secrets.LocalDevProject,
		Name:    "SG_REPOFEREE_HMAC_KEY",
	}
	values.HMAC, err = store.GetExternal(ctx, sec)
	if err != nil {
		return values, err
	}

	return values, nil
}

func newRepofereeReq(ctx context.Context, endpoint, path string) (*http.Request, error) {
	urlPath, err := url.JoinPath(endpoint, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return nil, err
	}

	secrets, err := getRepofereeSecrets(ctx)
	if err != nil {
		return nil, err
	}
	authorization := fmt.Sprintf("Bearer %s", secrets.AuthToken)
	timestamp := time.Now().Format(time.RFC3339)

	signature := hmacSign(secrets.HMAC, fmt.Sprintf("%s:%s:%s", timestamp, path, authorization))
	req.Header.Add("Authorization", authorization)
	req.Header.Add("X-Timestamp", timestamp)
	req.Header.Add("X-Signature", hex.EncodeToString(signature))
	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

func getRepofereeReport(ctx *cli.Context) error {
	pending := std.Out.Pending(output.Line("⌛", output.StylePending, "Fetching latest repoferee report"))

	pending.Update("Creating signed request")
	req, err := newRepofereeReq(ctx.Context, RepofereeEndpoint, "/results")
	if err != nil {
		pending.Complete(output.Line("❌", output.StyleFailure, "Failed to create signed request"))
		return err
	}

	pending.Updatef("Making request to %q", req.URL.String())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		pending.Complete(output.Line("❌", output.StyleFailure, "Request failed"))
		return err
	}
	defer resp.Body.Close()
	pending.WriteLine(output.Line("✅", output.StyleSuccess, "Response received from repoferee"))

	dst := os.Stderr
	if ctx.String("output-file") != "" {
		fd, err := os.Create(ctx.String("output-file"))
		if err != nil {
			pending.Complete(output.Line("❌", output.StyleFailure, "Failed to open output file %q"))
			return err
		}
		dst = fd
	}
	pending.Update("Decoding and formatting response")
	value := struct {
		Results map[string]json.RawMessage `json:"results"`
		LastRun string                     `json:"last_run_at"`
		NextRun string                     `json:"next_run_at"`
	}{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&value); err != nil {
		pending.Complete(output.Line("❌", output.StyleFailure, "Failed to decode response"))
		return err
	}

	filename := "stderr"
	if dst != os.Stderr {
		filename = dst.Name()
	}
	enc := json.NewEncoder(dst)
	enc.SetIndent(" ", " ")
	if err := enc.Encode(value); err != nil {
		pending.Complete(output.Linef("❌", output.StyleFailure, "Failed to write to %q", filename))
		return err
	}
	pending.Complete(output.Linef("✅", output.StyleSuccess, "Report written to %q", filename))
	std.Out.WriteMarkdown(fmt.Sprintf("Last report run: `%s`\nNext report run: `%s`\n", value.LastRun, value.NextRun))
	return nil
}
