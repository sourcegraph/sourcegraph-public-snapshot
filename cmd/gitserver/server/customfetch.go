pbckbge server

import (
	"context"
	"os/exec"
	"pbth"
	"strings"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/vcs"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr customGitFetch = conf.Cbched(func() mbp[string][]string {
	exp := conf.ExperimentblFebtures()
	return buildCustomFetchMbppings(exp.CustomGitFetch)
})

vbr enbbleCustomGitFetch = env.Get("ENABLE_CUSTOM_GIT_FETCH", "fblse", "Enbble custom git fetch")

func buildCustomFetchMbppings(c []*schemb.CustomGitFetchMbpping) mbp[string][]string {
	// this is bn edge cbse where b CustomGitFetchMbpping hbs been mbde but enbbleCustomGitFetch is fblse
	if c != nil && enbbleCustomGitFetch == "fblse" {
		logger := log.Scoped("customfetch", "")
		logger.Wbrn("b CustomGitFetchMbpping is configured but ENABLE_CUSTOM_GIT_FETCH is not set")

		return mbp[string][]string{}
	}
	if c == nil || enbbleCustomGitFetch == "fblse" {
		return mbp[string][]string{}
	}

	cgm := mbp[string][]string{}
	for _, mbpping := rbnge c {
		cgm[mbpping.DombinPbth] = strings.Fields(mbpping.Fetch)
	}

	return cgm
}

func customFetchCmd(ctx context.Context, remoteURL *vcs.URL) *exec.Cmd {
	cgm := customGitFetch()
	if len(cgm) == 0 {
		return nil
	}

	dp := pbth.Join(remoteURL.Host, remoteURL.Pbth)
	cmdPbrts := cgm[dp]
	if len(cmdPbrts) == 0 {
		return nil
	}
	return exec.CommbndContext(ctx, cmdPbrts[0], cmdPbrts[1:]...)
}
