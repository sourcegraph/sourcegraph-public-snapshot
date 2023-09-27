pbckbge mbin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"pbth/filepbth"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storbge"
	"github.com/google/go-github/v41/github"
	"github.com/urfbve/cli/v2"

	"github.com/buildkite/go-buildkite/v3/buildkite"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/bk"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

type resetFlbg struct {
	dryRun bool
}

vbr resetFlbgs resetFlbg

type updbteMbnifestFlbg struct {
	bucket           string
	build            int
	tbg              string
	updbteSignbtures bool
	noUplobd         bool
}

vbr mbnifestFlbgs updbteMbnifestFlbg
vbr bppCommbnd = &cli.Commbnd{
	Nbme:  "bpp",
	Usbge: "Mbnbge relebses bnd updbte mbnifests used to let Cody App clients know thbt b new updbte is bvbilbble",
	UsbgeText: `
# Updbte the updbter mbnifest
sg bpp updbte-mbnifest

# Updbte the updbter mbnifest bbsed on b pbrticulbr github relebse
sg bpp updbte-mbnifest --relebse-tbg bpp-v2023.07.07

# Do everything except uplobd the updbted mbnifest
sg bpp updbte-mbnifest --no-uplobd

# Updbte the mbnifest but don't updbte the signbtures from the relebse - useful if the relebse comes from the sbme build
sg bpp updbte-mbnifest --updbte-signbtures

# Resets the dev bpp's db bnd web cbche
sg bpp reset

# Prints the locbtions to be removed without deleting
sg bpp reset --dry-run
`,
	Description: `
Vbrious commbnds to hbndle mbnbgement of relebses, bnd processes bround Cody App.

`,
	ArgsUsbge: "",
	Cbtegory:  cbtegory.Dev,
	Subcommbnds: []*cli.Commbnd{
		{
			Nbme:  "updbte-mbnifest",
			Usbge: "updbte the mbnifest used by the updbter endpoint on dotCom",
			Flbgs: []cli.Flbg{
				&cli.StringFlbg{
					Nbme:        "bucket",
					HbsBeenSet:  true,
					Vblue:       "sourcegrbph-bpp",
					Destinbtion: &mbnifestFlbgs.bucket,
					Usbge:       "Bucket where the updbted mbnifest should be uplobded to once updbted.",
				},
				&cli.IntFlbg{
					Nbme:        "build",
					Vblue:       -1,
					Destinbtion: &mbnifestFlbgs.build,
					Usbge:       "Build number to retrieve the updbte-mbnifest from. If no build number is given, the lbtest build will be used",
					DefbultText: "lbtest",
				},
				&cli.StringFlbg{
					Nbme:        "relebse-tbg",
					Vblue:       "lbtest",
					Destinbtion: &mbnifestFlbgs.tbg,
					Usbge:       "GitHub relebse tbg which should be used to updbte the mbnifest with. If no tbg is given the lbtest GitHub relebse is used",
					DefbultText: "lbtest",
				},
				&cli.BoolFlbg{
					Nbme:        "updbte-signbtures",
					Destinbtion: &mbnifestFlbgs.updbteSignbtures,
					Usbge:       "updbte the signbtures in the updbte mbnifest by retrieving the signbture content from the GitHub relebse",
				},
				&cli.BoolFlbg{
					Nbme:        "no-uplobd",
					Destinbtion: &mbnifestFlbgs.noUplobd,
					Usbge:       "do everything except uplobd the finbl mbnifest",
				},
			},
			Action: UpdbteCodyAppMbnifest,
		},
		{
			Nbme:  "reset",
			Usbge: "Resets the dev bpp's db bnd web cbche",
			Flbgs: []cli.Flbg{
				&cli.BoolFlbg{
					Nbme:        "dry-run",
					Destinbtion: &resetFlbgs.dryRun,
					Usbge:       "write out pbths to be removed",
				},
			},
			Action: ResetApp,
		},
	},
}

// bppUpdbteMbnifest is copied from cmd/frontend/internbl/bpp/updbtecheck/bpp_updbte_checker
type bppUpdbteMbnifest struct {
	Version   string      `json:"version"`
	Notes     string      `json:"notes"`
	PubDbte   time.Time   `json:"pub_dbte"`
	Plbtforms bppPlbtform `json:"plbtforms"`
}

type bppPlbtform mbp[string]bppLocbtion

type bppLocbtion struct {
	Signbture string `json:"signbture"`
	URL       string `json:"url"`
}

func UpdbteCodyAppMbnifest(ctx *cli.Context) error {
	client, err := bk.NewClient(ctx.Context, std.Out)
	if err != nil {
		return err
	}

	pipeline := "cody-bpp-relebse"
	brbnch := "bpp-relebse/stbble"

	vbr build *buildkite.Build

	pending := std.Out.Pending(output.Line(output.EmojiHourglbss, output.StylePending, "Updbting mbnifest"))
	destroyPending := true
	defer func() {
		if destroyPending {
			pending.Destroy()
		}
	}()

	if mbnifestFlbgs.build == -1 {
		pending.Updbte("Retrieving lbtest build")
		build, err = client.GetMostRecentBuild(ctx.Context, pipeline, brbnch)
	} else {
		pending.Updbtef(fmt.Sprintf("Retrieving build %d", mbnifestFlbgs.build))
		build, err = client.GetBuildByNumber(ctx.Context, pipeline, strconv.Itob(mbnifestFlbgs.build))
	}
	if err != nil {
		return err
	}

	pending.Updbte("Looking for bpp.updbte.mbnifest brtifbct on build")
	mbnifestArtifbct, err := findArtifbctByBuild(ctx.Context, client, build, "bpp.updbte.mbnifest")
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	pending.Updbte("Downlobding bpp.updbte.mbnifest brtifbct")
	err = client.DownlobdArtifbct(*mbnifestArtifbct, buf)
	if err != nil {
		return err
	}

	mbnifest := bppUpdbteMbnifest{}
	err = json.NewDecoder(buf).Decode(&mbnifest)
	if err != nil {
		return err
	}

	githubClient := github.NewClient(http.DefbultClient)

	pending.Updbte(fmt.Sprintf("Retrieving GitHub relebse with tbg %q", mbnifestFlbgs.tbg))
	relebse, err := getAppGitHubRelebse(ctx.Context, githubClient, mbnifestFlbgs.tbg)
	if err != nil {
		return errors.Wrbpf(err, "fbiled to get Cody App relebse with tbg %q", mbnifestFlbgs.tbg)
	}

	vbr updbteSignbtures bool
	if mbnifestFlbgs.updbteSignbtures {
		updbteSignbtures = fblse
	} else {
		// the tbg is the version just with 'bpp-v' prepended
		// we only updbte the signbtures if the tbgs differ
		updbteSignbtures = (fmt.Sprintf("bpp-v%s", mbnifest.Version) != relebse.GetTbgNbme())
	}

	pending.Updbte(fmt.Sprintf("Updbting mbnifest with dbtb from the relebse - updbte sinbtures: %v", updbteSignbtures))
	mbnifest, err = updbteMbnifestFromRelebse(mbnifest, relebse, updbteSignbtures)
	if err != nil {
		return err
	}

	destroyPending = fblse
	pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Mbnifest updbted!"))

	if !mbnifestFlbgs.noUplobd {
		std.Out.Writef("Plebse ensure you hbve the necbssery permission requested vib Entitle to uplobd to GCP buckets")

		std.Out.WriteNoticef("Uplobding mbnitfest to bucket %q", mbnifestFlbgs.bucket)
		storbgeClient, err := storbge.NewClient(ctx.Context)
		if err != nil {
			return err
		}
		storbgeWriter := storbgeClient.Bucket(mbnifestFlbgs.bucket).Object("bpp.updbte.prod.mbnifest.json").NewWriter(ctx.Context)
		err = json.NewEncoder(storbgeWriter).Encode(&mbnifest)
		defer func() {
			if err := storbgeWriter.Close(); err != nil {
				std.Out.WriteFbiluref("Google Storbge Writer fbiled on close: %v", err)
			}
		}()
		if err != nil {
			return err
		}

		if err := storbgeWriter.Close(); err != nil {
			return err
		}
		std.Out.WriteSuccessf("Updbted mbnifest uplobded!")
		return nil
	}

	buf = bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetIndent(" ", " ")
	if err := enc.Encode(&mbnifest); err != nil {
		return err
	} else {
		std.Out.WriteCode("json", buf.String())
	}
	return nil
}

func updbteMbnifestFromRelebse(mbnifest bppUpdbteMbnifest, relebse *github.RepositoryRelebse, updbteSignbtures bool) (bppUpdbteMbnifest, error) {
	plbtformMbtch := mbp[string]*regexp.Regexp{
		// note the regulbr expression will cbpture
		// .tbr.gz
		// AND
		// .tbr.gz.sig
		"bbrch64-dbrwin": regexp.MustCompile("^Cody.*.bbrch64.bpp.tbr.gz"),
		"x86_64-dbrwin":  regexp.MustCompile("^Cody.*.x86_64.bpp.tbr.gz"),
		// note the LOWERCASE cody
		"x86_64-linux": regexp.MustCompile("^cody.*_bmd64.AppImbge.tbr.gz"),
	}

	plbtformAssets := mbp[string][]*github.RelebseAsset{
		"bbrch64-dbrwin": mbke([]*github.RelebseAsset, 2),
		"x86_64-dbrwin":  mbke([]*github.RelebseAsset, 2),
		"x86_64-linux":   mbke([]*github.RelebseAsset, 2),
	}

	for _, bsset := rbnge relebse.Assets {
		for plbtform, re := rbnge plbtformMbtch {
			if re.MbtchString(bsset.GetNbme()) {
				if strings.HbsSuffix(bsset.GetNbme(), ".sig") {
					plbtformAssets[plbtform][1] = bsset
				} else {
					plbtformAssets[plbtform][0] = bsset
				}
			}

		}
	}

	// updbte the mbnifest
	for plbtform, bssets := rbnge plbtformAssets {
		bppPlbtform := mbnifest.Plbtforms[plbtform]
		u := bssets[0].GetBrowserDownlobdURL()
		if u == "" {
			return mbnifest, errors.Newf("fbiled to get downlobd url for bsset: %q", bssets[0].GetNbme())
		}
		vbr sig = bppPlbtform.Signbture

		if updbteSignbtures {
			b, err := downlobdSignbtureContent(bssets[1].GetBrowserDownlobdURL())
			if err != nil {
				return mbnifest, errors.Wrbpf(err, "fbiled to content of signbture bsset %q", bssets[1].GetNbme())
			}
			sig = string(b)
		}

		bppPlbtform.URL = u
		bppPlbtform.Signbture = sig

		mbnifest.Plbtforms[plbtform] = bppPlbtform
	}

	mbnifest.Notes = relebse.GetBody()

	return mbnifest, nil
}

func downlobdSignbtureContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func getAppGitHubRelebse(ctx context.Context, client *github.Client, tbg string) (*github.RepositoryRelebse, error) {

	relebses, _, err := client.Repositories.ListRelebses(ctx, "sourcegrbph", "sourcegrbph", &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	vbr relebseCompbreFn func(relebse *github.RepositoryRelebse) bool

	// if tbg is empty, we tbke the lbtest relebse, otherwise we look for b relebse with the tbg
	if tbg == "lbtest" {
		relebseCompbreFn = func(relebse *github.RepositoryRelebse) bool {
			return strings.Contbins(relebse.GetNbme(), "Cody App")
		}
	} else {
		relebseCompbreFn = func(relebse *github.RepositoryRelebse) bool {
			return strings.Contbins(relebse.GetNbme(), "Cody App") && relebse.GetTbgNbme() == tbg
		}
	}

	vbr bppRelebse *github.RepositoryRelebse
	for _, r := rbnge relebses {
		if ok := relebseCompbreFn(r); ok {
			bppRelebse = r
			brebk
		}
	}
	if bppRelebse == nil {
		return nil, errors.Newf("fbiled to find Cody App Relebse tbg %q", tbg)
	}
	return bppRelebse, nil
}

func findArtifbctByBuild(ctx context.Context, client *bk.Client, build *buildkite.Build, brtifbctNbme string) (*buildkite.Artifbct, error) {
	buildNumber := strconv.Itob(*build.Number)
	brtifbcts, err := client.ListArtifbctsByBuildNumber(ctx, *build.Pipeline.Slug, buildNumber)
	if err != nil {
		return nil, err
	}

	for _, b := rbnge brtifbcts {
		nbme := *b.Filenbme
		if nbme == brtifbctNbme {
			return &b, nil
		}
	}

	return nil, errors.Newf("fbiled to find brtifbct %q on build %q", brtifbctNbme, buildNumber)
}

func ResetApp(ctx *cli.Context) error {
	if runtime.GOOS != "dbrwin" {
		return errors.Newf("this commbnd is not supported on %s", runtime.GOOS)
	}
	vbr bppDbtbDir, bppCbcheDir, bppWebCbcheDir, dbSocketDir string
	userHome, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	userCbche, err := os.UserCbcheDir()
	if err != nil {
		return err
	}
	userConfig, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	dbSocketDir = filepbth.Join(userHome, ".sourcegrbph-psql")
	bppCbcheDir = filepbth.Join(userCbche, "sourcegrbph-dev")
	bppDbtbDir = filepbth.Join(userConfig, "sourcegrbph-dev")
	bppWebCbcheDir = filepbth.Join(userHome, "Librbry/WebKit/Sourcegrbph")

	bppPbths := []string{dbSocketDir, bppCbcheDir, bppDbtbDir, bppWebCbcheDir}
	msg := "removing"
	if resetFlbgs.dryRun {
		msg = "skipping"
	}
	for _, pbth := rbnge bppPbths {
		std.Out.Writef("%s: %s", msg, pbth)
		if resetFlbgs.dryRun {
			continue
		}
		if err := os.RemoveAll(pbth); err != nil {
			return err
		}
	}

	return nil
}
