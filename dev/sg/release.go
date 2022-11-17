package main

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	regexp "github.com/grafana/regexp"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/bk"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var releaseCommand = &cli.Command{
	Name:     "release",
	Usage:    "Sourcegraph release utilities",
	Category: CategoryUtil,
	Subcommands: []*cli.Command{{
		Name:     "cve-check",
		Usage:    "Check all CVEs found in a buildkite build against a set of preapproved CVEs for a release",
		Category: CategoryUtil,
		Action:   cveCheck,
		Flags: []cli.Flag{
			&buildNumberFlag,
			&referenceUriFlag,
		},
		UsageText: `sg release cve-check -u https://handbook.sourcegraph.com/departments/security/tooling/trivy/4-2-0/ -b 184191`,
	}},
}

var buildNumberFlag = cli.StringFlag{
	Name:     "buildNumber",
	Usage:    "The buildkite build number to check for CVEs",
	Required: true,
	Aliases:  []string{"b"},
}

var referenceUriFlag = cli.StringFlag{
	Name:     "uri",
	Usage:    "A reference url that contains approved CVEs. Often a link to a handbook page eg: https://handbook.sourcegraph.com/departments/security/tooling/trivy/4-2-0/.",
	Required: true,
	Aliases:  []string{"u"},
}

func cveCheck(cmd *cli.Context) error {
	std.Out.WriteLine(output.Styledf(output.StylePending, "Checking release for approved CVEs..."))

	referenceUrl := referenceUriFlag.Get(cmd)
	number := buildNumberFlag.Get(cmd)

	client, err := bk.NewClient(cmd.Context, std.Out)
	if err != nil {
		return errors.Wrap(err, "bk.NewClient")
	}

	artifacts, err := client.ListArtifactsByBuildNumber(cmd.Context, "sourcegraph", number)
	if err != nil {
		return errors.Wrap(err, "unable to list artifacts by build number")
	}

	var hbBuf bytes.Buffer
	err = downloadUrl(referenceUrl, &hbBuf)
	if err != nil {
		return errors.Wrap(err, "unable to download reference uri")
	}
	hbPage := hbBuf.String()
	if len(hbPage) == 0 {
		return errors.New("provided reference uri does not have any contents")
	}

	var foundCVE []string
	var unapprovedCVE []string

	for _, artifact := range artifacts {
		name := *artifact.Filename
		url := *artifact.DownloadURL

		pattern, err := regexp.Compile(`<\w+>(CVE.*)<\/\w+>`)
		if err != nil {
			return errors.Wrap(err, "failed to build regexp pattern")
		}

		if strings.HasSuffix(*artifact.Filename, "security-report.html") {
			std.Out.WriteLine(output.Styledf(output.StylePending, "Checking security report: %s %s", name, url))

			var buf bytes.Buffer
			err = client.DownloadArtifact(artifact, &buf)
			if err != nil {
				return errors.Newf("failed to download security artifact %q at %s: %w", name, url, err)
			}

			matches := pattern.FindAllStringSubmatch(buf.String(), -1)
			for _, match := range matches {
				cve := strings.TrimSpace(match[1])
				foundCVE = append(foundCVE, cve)
				if !strings.Contains(hbPage, cve) {
					unapprovedCVE = append(unapprovedCVE, cve)
				}
			}
		}
	}

	// verbose := verboseFlag.Get(cmd)
	std.Out.WriteLine(output.Styledf(output.StyleBold, "Found %d CVEs in the build", len(foundCVE)))
	if verbose {
		for _, s := range foundCVE {
			std.Out.WriteLine(output.Styledf(output.StyleWarning, "%s", s))
		}
	}
	if len(unapprovedCVE) > 0 {
		std.Out.WriteLine(output.Styledf(output.StyleWarning, "Unable to match CVEs"))
		for _, s := range unapprovedCVE {
			std.Out.WriteLine(output.Styledf(output.StyleWarning, "%s", s))
		}
	} else {
		std.Out.WriteLine(output.Styledf(output.StyleSuccess, "All CVEs approved!"))
	}

	return nil
}

func downloadUrl(uri string, w io.Writer) (err error) {
	std.Out.WriteLine(output.Styledf(output.StylePending, "Downloading url: %s", uri))
	client := http.Client{}
	resp, err := client.Get(uri)
	if err != nil {
		return err
	}
	defer func() { err = resp.Body.Close() }()

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
