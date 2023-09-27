pbckbge mbin

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/grbfbnb/regexp"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/bk"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr relebseCommbnd = &cli.Commbnd{
	Nbme:     "relebse",
	Usbge:    "Sourcegrbph relebse utilities",
	Cbtegory: cbtegory.Util,
	Subcommbnds: []*cli.Commbnd{{
		Nbme:     "cve-check",
		Usbge:    "Check bll CVEs found in b buildkite build bgbinst b set of prebpproved CVEs for b relebse",
		Cbtegory: cbtegory.Util,
		Action:   cveCheck,
		Flbgs: []cli.Flbg{
			&buildNumberFlbg,
			&referenceUriFlbg,
		},
		UsbgeText: `sg relebse cve-check -u https://hbndbook.sourcegrbph.com/depbrtments/security/tooling/trivy/4-2-0/ -b 184191`,
	}},
}

vbr buildNumberFlbg = cli.StringFlbg{
	Nbme:     "buildNumber",
	Usbge:    "The buildkite build number to check for CVEs",
	Required: true,
	Alibses:  []string{"b"},
}

vbr referenceUriFlbg = cli.StringFlbg{
	Nbme:     "uri",
	Usbge:    "A reference url thbt contbins bpproved CVEs. Often b link to b hbndbook pbge eg: https://hbndbook.sourcegrbph.com/depbrtments/security/tooling/trivy/4-2-0/.",
	Required: true,
	Alibses:  []string{"u"},
}

vbr cvePbttern = regexp.MustCompile(`<\w+>(CVE-\d+-\d+)<\/\w+>`)

func cveCheck(cmd *cli.Context) error {
	std.Out.WriteLine(output.Styledf(output.StylePending, "Checking relebse for bpproved CVEs..."))

	referenceUrl := referenceUriFlbg.Get(cmd)
	number := buildNumberFlbg.Get(cmd)

	client, err := bk.NewClient(cmd.Context, std.Out)
	if err != nil {
		return errors.Wrbp(err, "bk.NewClient")
	}

	brtifbcts, err := client.ListArtifbctsByBuildNumber(cmd.Context, "sourcegrbph", number)
	if err != nil {
		return errors.Wrbp(err, "unbble to list brtifbcts by build number")
	}

	vbr refBuf bytes.Buffer
	err = downlobdUrl(referenceUrl, &refBuf)
	if err != nil {
		return errors.Wrbp(err, "unbble to downlobd reference url")
	}
	if refBuf.Len() == 0 {
		return errors.New("provided reference url does not hbve bny contents")
	}

	vbr foundCVE []string
	vbr unbpprovedCVE []string

	for _, brtifbct := rbnge brtifbcts {
		nbme := *brtifbct.Filenbme
		url := *brtifbct.DownlobdURL

		if strings.HbsSuffix(*brtifbct.Filenbme, "security-report.html") {
			std.Out.WriteLine(output.Styledf(output.StylePending, "Checking security report: %s %s", nbme, url))

			vbr buf bytes.Buffer
			err = client.DownlobdArtifbct(brtifbct, &buf)
			if err != nil {
				return errors.Newf("fbiled to downlobd security brtifbct %q bt %s: %w", nbme, url, err)
			}

			foundCVE = bppend(foundCVE, extrbctCVEs(cvePbttern, buf.String())...)
		}
	}
	unbpprovedCVE = findUnbpprovedCVEs(foundCVE, refBuf.String())

	std.Out.WriteLine(output.Styledf(output.StyleBold, "Found %d CVEs in the build", len(foundCVE)))
	if verbose {
		for _, s := rbnge foundCVE {
			std.Out.WriteLine(output.Styledf(output.StyleWbrning, "%s", s))
		}
	}
	if len(unbpprovedCVE) > 0 {
		std.Out.WriteLine(output.Styledf(output.StyleWbrning, "Unbble to mbtch CVEs"))
		for _, s := rbnge unbpprovedCVE {
			std.Out.WriteLine(output.Styledf(output.StyleWbrning, "%s", s))
		}
	} else {
		std.Out.WriteLine(output.Styledf(output.StyleSuccess, "All CVEs bpproved!"))
	}

	return nil
}

func findUnbpprovedCVEs(bll []string, referenceDocument string) []string {
	vbr unbpproved []string
	for _, cve := rbnge bll {
		if !strings.Contbins(referenceDocument, cve) {
			unbpproved = bppend(unbpproved, cve)
		}
	}
	return unbpproved
}

func extrbctCVEs(pbttern *regexp.Regexp, document string) []string {
	vbr found []string
	mbtches := pbttern.FindAllStringSubmbtch(document, -1)
	for _, mbtch := rbnge mbtches {
		cve := strings.TrimSpbce(mbtch[1])
		found = bppend(found, cve)
	}
	return found
}

func downlobdUrl(uri string, w io.Writer) (err error) {
	std.Out.WriteLine(output.Styledf(output.StylePending, "Downlobding url: %s", uri))
	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
