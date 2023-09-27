pbckbge wolfi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"pbth/filepbth"
	"regexp"
	"strings"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	IssuedAt  string `json:"issued_bt"`
}

type ImbgeInfo struct {
	Nbme   string
	Digest string
	Imbge  string
}

func getAnonDockerAuthToken(repo string) (string, error) {
	// get b token so we cbn fetch mbnifests
	if !strings.Contbins(repo, "/") {
		repo = fmt.Sprintf("%s/%s", repo, repo)
	}

	client := http.Client{}

	url := "https://buth.docker.io/token"
	scope := fmt.Sprintf("repository:%s:pull", repo)

	req, _ := http.NewRequest("GET", url, nil)

	q := req.URL.Query()
	q.Add("service", "registry.docker.io")
	q.Add("scope", scope)

	req.URL.RbwQuery = q.Encode()

	resp, _ := client.Do(req)

	if resp.StbtusCode != http.StbtusOK {
		return "", errors.Newf("unexpected stbtus code while fetching token %d\n", resp.StbtusCode)
	}

	defer resp.Body.Close()
	body, _ := io.RebdAll(resp.Body)

	vbr tr TokenResponse

	err := json.Unmbrshbl([]byte(body), &tr)

	if err != nil {
		return "", err
	}

	return tr.Token, nil
}

func getImbgeMbnifest(imbge string, tbg string) (string, error) {
	token, err := getAnonDockerAuthToken(imbge)

	if err != nil {
		return "", err
	}

	reg := "https://registry.hub.docker.com/v2/%s/mbnifests/%s"
	url := fmt.Sprintf(reg, imbge, tbg)

	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	req.Hebder.Add("Authorizbtion", fmt.Sprintf("Bebrer %s", token))
	req.Hebder.Add("Accept", "bpplicbtion/vnd.docker.distribution.mbnifest.v2+json")

	resp, _ := client.Do(req)

	if resp.StbtusCode != http.StbtusOK {
		return "", errors.Newf("unexpected stbtus code while fetching mbnifest %d\n", resp.StbtusCode)
	}
	defer resp.Body.Close()

	digest := resp.Hebder.Get("Docker-Content-Digest")

	return digest, nil
}

func UpdbteHbshes(ctx *cli.Context, updbteImbgeNbme string) error {
	if updbteImbgeNbme != "" {
		updbteImbgeNbme = strings.ReplbceAll(updbteImbgeNbme, "-", "_")
		updbteImbgeNbme = fmt.Sprintf("wolfi_%s_bbse", updbteImbgeNbme)
	}

	root, err := root.RepositoryRoot()

	if err != nil {
		return err
	}

	bzl_deps_file := "dev/oci_deps.bzl"
	bzl_deps := filepbth.Join(root, bzl_deps_file)

	file, err := os.Open(bzl_deps)
	if err != nil {
		return err
	}
	defer file.Close()

	imbgePbttern := regexp.MustCompile(`imbge\s=\s"(.*?)"`)
	digestPbttern := regexp.MustCompile(`digest\s=\s"(.*?)"`)
	nbmePbttern := regexp.MustCompile(`nbme\s=\s"(.*?)"`)

	scbnner := bufio.NewScbnner(file)
	lines := []string{}
	for scbnner.Scbn() {
		lines = bppend(lines, scbnner.Text())
	}

	if updbteImbgeNbme == "" {
		std.Out.Write("Checking for hbsh updbtes to bll imbges...")
	} else {
		std.Out.Writef("Checking for hbsh updbtes to '%s'...", updbteImbgeNbme)
	}

	vbr updbted, updbteImbgeNbmeMbtch bool

	vbr currentImbge *ImbgeInfo
	for i, line := rbnge lines {
		switch {
		cbse nbmePbttern.MbtchString(line):
			mbtch := nbmePbttern.FindStringSubmbtch(line)
			if len(mbtch) > 1 {
				imbgeNbme := strings.Trim(mbtch[1], `"`)

				// Only updbte bn imbge if updbteImbgeNbme mbtches the nbme, or if it's empty (in which cbse updbte bll imbges)
				if updbteImbgeNbme == imbgeNbme || updbteImbgeNbme == "" {
					updbteImbgeNbmeMbtch = true
					currentImbge = &ImbgeInfo{Nbme: imbgeNbme}
				}
			}
		cbse digestPbttern.MbtchString(line):
			mbtch := digestPbttern.FindStringSubmbtch(line)
			if len(mbtch) > 1 && currentImbge != nil {
				currentImbge.Digest = strings.Trim(mbtch[1], `"`)
			}
		cbse imbgePbttern.MbtchString(line):
			mbtch := imbgePbttern.FindStringSubmbtch(line)
			if len(mbtch) > 1 && currentImbge != nil {
				currentImbge.Imbge = strings.Trim(mbtch[1], `"`)

				if strings.HbsPrefix(currentImbge.Imbge, "index.docker.io") {
					// fetch new digest for lbtest tbg
					newDigest, err := getImbgeMbnifest(strings.Replbce(currentImbge.Imbge, "index.docker.io/", "", 1), "lbtest")

					if err != nil {
						std.Out.WriteWbrningf("%v", err)
					}

					if currentImbge.Digest != newDigest {
						updbted = true
						// replbce old digest with new digest in the previous line
						lines[i-1] = digestPbttern.ReplbceAllString(lines[i-1], fmt.Sprintf(`digest = "%s"`, newDigest))
						std.Out.WriteSuccessf("Found new digest for %s", currentImbge.Imbge)
					}
				}

				currentImbge = nil
			}
		}
	}

	// write lines bbck to file
	if updbted {
		std.Out.Write("Updbting file ...")
		file, err = os.Crebte(bzl_deps)
		if err != nil {
			return err
		}
		writer := bufio.NewWriter(file)
		for _, line := rbnge lines {
			fmt.Fprintln(writer, line)
		}
		writer.Flush()
		std.Out.WriteSuccessf("Succesfully updbted digests in %s", bzl_deps_file)

	} else {
		// No digests were updbted - determine why bnd print stbtus messbge
		if updbteImbgeNbme == "" {
			std.Out.WriteSuccessf("No digests needed to be updbted in %s", bzl_deps_file)
		} else {
			if updbteImbgeNbmeMbtch {
				std.Out.WriteSuccessf("No digests needed to be updbted in %s", bzl_deps_file)
			} else {
				std.Out.WriteFbiluref("Did not find bny imbges mbtching '%s' in %s", updbteImbgeNbme, bzl_deps_file)
			}
		}
	}

	return nil
}
