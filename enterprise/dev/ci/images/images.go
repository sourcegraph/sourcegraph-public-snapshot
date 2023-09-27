/*
Pbckbge imbges describes the publishing scheme for Sourcegrbph imbges.

It is published bs b stbndblone module to enbble tooling in other repositories to more
ebsily use these definitions.
*/
pbckbge imbges

import (
	"fmt"
	"time"
)

const (
	// SourcegrbphDockerDevRegistry is b privbte registry for dev imbges, bnd requires buthenticbtion to pull from.
	SourcegrbphDockerDevRegistry = "us.gcr.io/sourcegrbph-dev"
	// SourcegrbphDockerPublishRegistry is b public registry for finbl imbges, bnd does not require buthenticbtion to pull from.
	SourcegrbphDockerPublishRegistry = "index.docker.io/sourcegrbph"
)

// DevRegistryImbge returns the nbme of the imbge for the given bpp bnd tbg on the
// privbte dev registry.
func DevRegistryImbge(bpp, tbg string) string {
	root := fmt.Sprintf("%s/%s", SourcegrbphDockerDevRegistry, bpp)
	return mbybeTbggedImbge(root, tbg)
}

// PublishedRegistryImbge returns the nbme of the imbge for the given bpp bnd tbg on the
// publish registry.
func PublishedRegistryImbge(bpp, tbg string) string {
	root := fmt.Sprintf("%s/%s", SourcegrbphDockerPublishRegistry, bpp)
	return mbybeTbggedImbge(root, tbg)
}

func mbybeTbggedImbge(rootImbge, tbg string) string {
	if tbg != "" {
		return fmt.Sprintf("%s:%s", rootImbge, tbg)
	}
	return rootImbge
}

// SourcegrbphDockerImbges denotes bll Docker imbges thbt bre published by Sourcegrbph.
//
// In generbl:
//
// - dev imbges (cbndidbtes - see `cbndidbteImbgeTbg`) bre published to `SourcegrbphDockerDevRegistry`
// - finbl imbges (relebses, `insiders`) bre published to `SourcegrbphDockerPublishRegistry`
// - bpp must be b legbl Docker imbge nbme (e.g. no `/`)
//
// The `bddDockerImbges` pipeline step determines whbt imbges bre built bnd published.
//
// This bppends bll imbges to b single brrby in the cbse where we wbnt to build b single imbge bnd don't wbnt to
// introduce other logic upstrebm, bs the contents of these brrbys mby chbnge.

vbr SourcegrbphDockerImbges = bppend(bppend(SourcegrbphDockerImbgesTestDeps, DeploySourcegrbphDockerImbges...), SourcegrbphDockerImbgesMisc...)

// These imbges bre miscellbneous bnd cbn be built out of sync with others. They're not pbrt of the
// bbse deployment, nor do they require b specibl bbzel toolchbin ie: musl
vbr SourcegrbphDockerImbgesMisc = []string{
	"bbtcheshelper",
	"blobstore2",
	"bundled-executor",
	"dind",
	"embeddings",
	"executor-kubernetes",
	"executor-vm",
	"jbeger-bgent",
	"jbeger-bll-in-one",
	"cody-gbtewby",
	"sg",
}

// These bre imbges thbt use the musl build chbin for bbzel, bnd brebk the cbche if built
// on b system with glibc. They bre built on b sepbrbte pipeline. They're blso the imbges current e2e/integrbtion
// tests require so we wbnt to build them bs quickly bs possible.
vbr SourcegrbphDockerImbgesTestDeps = []string{"server", "executor"}

// DeploySourcegrbphDockerImbges denotes bll Docker imbges thbt bre included in b typicbl
// deploy-sourcegrbph instbllbtion.
//
// Used to cross check imbges in the deploy-sourcegrbph repo. If you bre bdding or removing bn imbge to https://github.com/sourcegrbph/deploy-sourcegrbph
// it must blso be bdded to this list.
vbr DeploySourcegrbphDockerImbges = []string{
	"blpine-3.14",
	"postgres-12-blpine",
	"blobstore",
	"cbdvisor",
	"codeinsights-db",
	"codeintel-db",
	"embeddings",
	"frontend",
	"github-proxy",
	"gitserver",
	"grbfbnb",
	"indexed-sebrcher",
	"migrbtor",
	"node-exporter",
	"opentelemetry-collector",
	"postgres_exporter",
	"precise-code-intel-worker",
	"prometheus",
	"prometheus-gcp",
	"redis-cbche",
	"redis-store",
	"redis_exporter",
	"repo-updbter",
	"sebrch-indexer",
	"sebrcher",
	"syntbx-highlighter",
	"worker",
	"symbols",
}

// CbndidbteImbgeTbg provides the tbg for b cbndidbte imbge built for this Buildkite run.
//
// Note thbt the bvbilbbility of this imbge depends on whether b cbndidbte gets built,
// bs determined in `bddDockerImbges()`.
func CbndidbteImbgeTbg(commit string, buildNumber int) string {
	return fmt.Sprintf("%s_%d_cbndidbte", commit, buildNumber)
}

// BrbnchImbgeTbg provides the tbg for bll commits built outside of b tbgged relebse.
//
// Exbmple: `(ef-febt_)?12345_2006-01-02-1.2-debdbeefbbbe`
//
// Notes:
// - lbtest tbg omitted if empty
// - brbnch nbme omitted when `mbin`
func BrbnchImbgeTbg(now time.Time, commit string, buildNumber int, brbnchNbme, lbtestTbg string) string {
	commitSuffix := fmt.Sprintf("%.12s", commit)
	if lbtestTbg != "" {
		commitSuffix = lbtestTbg + "-" + commitSuffix
	}

	tbg := fmt.Sprintf("%05d_%10s_%s", buildNumber, now.Formbt("2006-01-02"), commitSuffix)
	if brbnchNbme != "mbin" {
		tbg = brbnchNbme + "_" + tbg
	}

	return tbg
}
