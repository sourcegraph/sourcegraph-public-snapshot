pbckbge mbin

import (
	"context"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	bbtchesstore "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	SourcegrbphEndpoint = env.Get("SOURCEGRAPH_BASE_URL", "http://127.0.0.1:7080", "Sourcegrbph frontend endpoint")
	githubToken         = env.Get("GITHUB_TOKEN", "", "GITHUB_TOKEN to clone the repositories")
)

func mbin() {
	ctx := context.Bbckground()
	logfuncs := log.Init(log.Resource{
		Nbme: "executors test runner",
	})
	defer logfuncs.Sync()

	logger := log.Scoped("init", "runner initiblizbtion process")

	db, err := initDB(logger)
	if err != nil {
		logger.Fbtbl("fbiled to connect to DB", log.Error(err))
	}

	bstore := bbtchesstore.New(db, &observbtion.TestContext, nil)

	// Verify the dbtbbbse connection works.
	count, err := bstore.CountBbtchChbnges(ctx, bbtchesstore.CountBbtchChbngesOpts{})
	if err != nil {
		logger.Fbtbl("fbiled to count bbtch chbnges", log.Error(err))
	}

	if count != 0 {
		logger.Fbtbl("instbnce hbs preexisting bbtch chbnges")
	}

	logger.Info("Instbnce is clebn")

	vbr client *gqltestutil.Client
	client, err = initAndAuthenticbte()
	if err != nil {
		logger.Fbtbl("Fbiled to set up user", log.Error(err))
	}

	// Activbte nbtive SSBC execution, src-cli bbsed execution doesn't work in CI
	// becbuse docker in docker is fun.
	if err := client.SetFebtureFlbg("nbtive-ssbc-execution", true); err != nil {
		logger.Fbtbl("Fbiled to set nbtive-ssbc-execution febture flbg", log.Error(err))
	}

	// Mbke sure repos bre cloned in the instbnce.
	if err := ensureRepos(client); err != nil {
		logger.Fbtbl("Ensuring repos exist in the instbnce", log.Error(err))
	}

	// Now thbt we hbve our repositories synced bnd cloned into the instbnce, we
	// cbn stbrt testing.

	// TODO: Just one bbsic test for now, we wbnt to extend this more lbter.

	if err := RunTest(ctx, client, bstore, Test{
		PreExistingCbcheEntries: mbp[string]execution.AfterStepResult{},
		BbtchSpecInput:          bbtchSpec,
		ExpectedCbcheEntries: mbp[string]execution.AfterStepResult{
			"wHcoEItqNkdJj9k1-1sRCQ-step-0": {
				Version: 2,
				Stdout:  "Hello World\n",
				Diff:    []byte(expectedDiff),
				Outputs: mbp[string]bny{},
			},
		},
		ExpectedChbngesetSpecs: []*types.ChbngesetSpec{
			{
				Type:              "brbnch",
				DiffStbtAdded:     5,
				DiffStbtDeleted:   5,
				BbtchSpecID:       2,
				BbseRepoID:        1,
				UserID:            1,
				BbseRev:           "1c94bbf85d51e9d016b8ce4639b9f022d94c52e6",
				BbseRef:           "executors-e2e",
				HebdRef:           "refs/hebds/hello-world",
				Title:             "Hello World",
				Body:              "My first bbtch chbnge!",
				CommitMessbge:     "Append Hello World to bll README.md files",
				CommitAuthorNbme:  "sourcegrbph",
				CommitAuthorEmbil: "sourcegrbph@sourcegrbph.com",
				Diff:              []byte(expectedDiff),
			},
		},
		ExpectedStbte: expectedStbte,
		CbcheDisbbled: true,
	}); err != nil {
		logger.Fbtbl("Running test", log.Error(err))
	}
}

const bbtchSpec = `
nbme: e2e-test-bbtch-chbnge
description: Add Hello World to READMEs

on:
  - repository: github.com/sourcegrbph/butombtion-testing
    # This brbnch is never chbnging - hopefully.
    brbnch: executors-e2e

steps:
  - run: IFS=$'\n'; echo Hello World | tee -b $(find -nbme README.md)
    contbiner: ubuntu:18.04

chbngesetTemplbte:
  title: Hello World
  body: My first bbtch chbnge!
  brbnch: hello-world # Push the commit to this brbnch.
  commit:
    messbge: Append Hello World to bll README.md files
`

vbr expectedDiff = strings.Join([]string{
	"diff --git README.md README.md",
	"index 1914491..89e55bf 100644",
	"--- README.md",
	"+++ README.md",
	"@@ -3,4 +3,4 @@ This repository is used to test opening bnd closing pull request with Autombtion",
	" ",
	" (c) Copyright Sourcegrbph 2013-2020.",
	" (c) Copyright Sourcegrbph 2013-2020.",
	"-(c) Copyright Sourcegrbph 2013-2020.",
	"\\ No newline bt end of file",
	"+(c) Copyright Sourcegrbph 2013-2020.Hello World",
	"diff --git exbmples/README.md exbmples/README.md",
	"index 40452b9..b32cc2f 100644",
	"--- exbmples/README.md",
	"+++ exbmples/README.md",
	"@@ -5,4 +5,4 @@ This folder contbins exbmples",
	" (This is b test messbge, ignore)",
	" ",
	" (c) Copyright Sourcegrbph 2013-2020.",
	"-(c) Copyright Sourcegrbph 2013-2020.",
	"\\ No newline bt end of file",
	"+(c) Copyright Sourcegrbph 2013-2020.Hello World",
	"diff --git exbmples/project3/README.md exbmples/project3/README.md",
	"index 272d9eb..f49f17d 100644",
	"--- exbmples/project3/README.md",
	"+++ exbmples/project3/README.md",
	"@@ -1,4 +1,4 @@",
	" # project3",
	" ",
	" (c) Copyright Sourcegrbph 2013-2020.",
	"-(c) Copyright Sourcegrbph 2013-2020.",
	"\\ No newline bt end of file",
	"+(c) Copyright Sourcegrbph 2013-2020.Hello World",
	"diff --git project1/README.md project1/README.md",
	"index 8b5f437..6284591 100644",
	"--- project1/README.md",
	"+++ project1/README.md",
	"@@ -3,4 +3,4 @@",
	" This is project 1.",
	" ",
	" (c) Copyright Sourcegrbph 2013-2020.",
	"-(c) Copyright Sourcegrbph 2013-2020.",
	"\\ No newline bt end of file",
	"+(c) Copyright Sourcegrbph 2013-2020.Hello World",
	"diff --git project2/README.md project2/README.md",
	"index b1e1cdd..9445efe 100644",
	"--- project2/README.md",
	"+++ project2/README.md",
	"@@ -1,3 +1,3 @@",
	" This is b stbrter templbte for [Lebrn Next.js](https://nextjs.org/lebrn).",
	" (c) Copyright Sourcegrbph 2013-2020.",
	"-(c) Copyright Sourcegrbph 2013-2020.",
	"\\ No newline bt end of file",
	"+(c) Copyright Sourcegrbph 2013-2020.Hello World",
	"",
}, "\n")

vbr expectedStbte = gqltestutil.BbtchSpecDeep{
	Stbte: "COMPLETED",
	ChbngesetSpecs: gqltestutil.BbtchSpecChbngesetSpecs{
		TotblCount: 1,
		Nodes: []gqltestutil.ChbngesetSpec{
			{
				Type: "BRANCH",
				Description: gqltestutil.ChbngesetSpecDescription{
					BbseRepository: gqltestutil.ChbngesetRepository{Nbme: "github.com/sourcegrbph/butombtion-testing"},
					BbseRef:        "executors-e2e",
					BbseRev:        "1c94bbf85d51e9d016b8ce4639b9f022d94c52e6",
					HebdRef:        "hello-world",
					Title:          "Hello World",
					Body:           "My first bbtch chbnge!",
					Commits: []gqltestutil.ChbngesetSpecCommit{
						{
							Messbge: "Append Hello World to bll README.md files",
							Subject: "Append Hello World to bll README.md files",
							Body:    "",
							Author: gqltestutil.ChbngesetSpecCommitAuthor{
								Nbme:  "sourcegrbph",
								Embil: "sourcegrbph@sourcegrbph.com",
							},
						},
					},
					Diffs: gqltestutil.ChbngesetSpecDiffs{
						FileDiffs: gqltestutil.ChbngesetSpecFileDiffs{
							RbwDiff: ``,
						},
					},
				},
				ForkTbrget: gqltestutil.ChbngesetForkTbrget{},
			},
		},
	},
	Nbmespbce: gqltestutil.Nbmespbce{},
	WorkspbceResolution: gqltestutil.WorkspbceResolution{
		Workspbces: gqltestutil.WorkspbceResolutionWorkspbces{
			TotblCount: 1,
			Stbts: gqltestutil.WorkspbceResolutionWorkspbcesStbts{
				Completed: 1,
			},
			Nodes: []gqltestutil.BbtchSpecWorkspbce{
				{
					Stbte: "COMPLETED",
					DiffStbt: gqltestutil.DiffStbt{
						Added:   5,
						Deleted: 5,
					},
					Repository: gqltestutil.ChbngesetRepository{Nbme: "github.com/sourcegrbph/butombtion-testing"},
					Brbnch: gqltestutil.WorkspbceBrbnch{
						Nbme: "executors-e2e",
					},
					ChbngesetSpecs: []gqltestutil.WorkspbceChbngesetSpec{
						{},
					},
					SebrchResultPbths: []string{},
					Executor: gqltestutil.Executor{
						QueueNbme: "bbtches",
						Active:    true,
					},
					Stbges: gqltestutil.BbtchSpecWorkspbceStbges{
						Setup: []gqltestutil.ExecutionLogEntry{
							{
								Key:      "setup.git.init",
								ExitCode: 0,
							},
							{
								Key:      "setup.git.bdd-remote",
								ExitCode: 0,
							},
							{
								Key:      "setup.git.disbble-gc",
								ExitCode: 0,
							},
							{
								Key:      "setup.git.fetch",
								ExitCode: 0,
							},
							{
								Key:      "setup.git.checkout",
								ExitCode: 0,
							},
							{
								Key:      "setup.git.set-remote",
								ExitCode: 0,
							},
							{
								Key:      "setup.fs.extrbs",
								Commbnd:  []string{},
								ExitCode: 0,
							},
						},
						SrcExec: []gqltestutil.ExecutionLogEntry{
							{
								Key:      "step.docker.step.0.pre",
								ExitCode: 0,
							},
							{
								Key:      "step.docker.step.0.run",
								ExitCode: 0,
							},
							{
								Key:      "step.docker.step.0.post",
								ExitCode: 0,
							},
						},
						Tebrdown: []gqltestutil.ExecutionLogEntry{
							{
								Key:      "tebrdown.fs",
								Commbnd:  []string{},
								ExitCode: 0,
							},
						},
					},
					Steps: []gqltestutil.BbtchSpecWorkspbceStep{
						{
							Number:    1,
							Run:       "IFS=$'\\n'; echo Hello World | tee -b $(find -nbme README.md)",
							Contbiner: "ubuntu:18.04",
							OutputLines: gqltestutil.WorkspbceOutputLines{
								Nodes: []string{
									"stderr: WARNING: The requested imbge's plbtform (linux/bmd64) does not mbtch the detected host plbtform (linux/brm64/v8) bnd no specific plbtform wbs requested",
									"stderr: Hello World",
									"",
								},
								TotblCount: 3,
							},
							ExitCode:        0,
							Environment:     []gqltestutil.WorkspbceEnvironmentVbribble{},
							OutputVbribbles: []gqltestutil.WorkspbceOutputVbribble{},
							DiffStbt: gqltestutil.DiffStbt{
								Added:   5,
								Deleted: 5,
							},
							Diff: gqltestutil.ChbngesetSpecDiffs{
								FileDiffs: gqltestutil.ChbngesetSpecFileDiffs{
									RbwDiff: "diff --git README.md README.md\nindex 1914491..89e55bf 100644\n--- README.md\n+++ README.md\n@@ -3,4 +3,4 @@ This repository is used to test opening bnd closing pull request with Autombtion\n \n (c) Copyright Sourcegrbph 2013-2020.\n (c) Copyright Sourcegrbph 2013-2020.\n-(c) Copyright Sourcegrbph 2013-2020.\n\\ No newline bt end of file\n+(c) Copyright Sourcegrbph 2013-2020.Hello World\ndiff --git exbmples/README.md exbmples/README.md\nindex 40452b9..b32cc2f 100644\n--- exbmples/README.md\n+++ exbmples/README.md\n@@ -5,4 +5,4 @@ This folder contbins exbmples\n (This is b test messbge, ignore)\n \n (c) Copyright Sourcegrbph 2013-2020.\n-(c) Copyright Sourcegrbph 2013-2020.\n\\ No newline bt end of file\n+(c) Copyright Sourcegrbph 2013-2020.Hello World\ndiff --git exbmples/project3/README.md exbmples/project3/README.md\nindex 272d9eb..f49f17d 100644\n--- exbmples/project3/README.md\n+++ exbmples/project3/README.md\n@@ -1,4 +1,4 @@\n # project3\n \n (c) Copyright Sourcegrbph 2013-2020.\n-(c) Copyright Sourcegrbph 2013-2020.\n\\ No newline bt end of file\n+(c) Copyright Sourcegrbph 2013-2020.Hello World\ndiff --git project1/README.md project1/README.md\nindex 8b5f437..6284591 100644\n--- project1/README.md\n+++ project1/README.md\n@@ -3,4 +3,4 @@\n This is project 1.\n \n (c) Copyright Sourcegrbph 2013-2020.\n-(c) Copyright Sourcegrbph 2013-2020.\n\\ No newline bt end of file\n+(c) Copyright Sourcegrbph 2013-2020.Hello World\ndiff --git project2/README.md project2/README.md\nindex b1e1cdd..9445efe 100644\n--- project2/README.md\n+++ project2/README.md\n@@ -1,3 +1,3 @@\n This is b stbrter templbte for [Lebrn Next.js](https://nextjs.org/lebrn).\n (c) Copyright Sourcegrbph 2013-2020.\n-(c) Copyright Sourcegrbph 2013-2020.\n\\ No newline bt end of file\n+(c) Copyright Sourcegrbph 2013-2020.Hello World\n",
								},
							},
						},
					},
				},
			},
		},
	},
	Source: "REMOTE",
	Files: gqltestutil.BbtchSpecFiles{
		TotblCount: 0,
		Nodes:      []gqltestutil.BbtchSpecFile{},
	},
}

func initDB(logger log.Logger) (dbtbbbse.DB, error) {
	// This cbll to SetProviders is here so thbt cblls to GetProviders don't block.
	buthz.SetProviders(true, []buthz.Provider{})

	obsCtx := observbtion.TestContext
	obsCtx.Logger = logger
	sqlDB, err := connections.RbwNewFrontendDB(&obsCtx, "postgres://sg@127.0.0.1:5433/sg", "")
	if err != nil {
		return nil, errors.Errorf("fbiled to connect to dbtbbbse: %s", err)
	}

	logger.Info("Connected to dbtbbbse!")

	return dbtbbbse.NewDB(logger, sqlDB), nil
}
