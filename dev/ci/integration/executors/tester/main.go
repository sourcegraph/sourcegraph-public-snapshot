package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

const currentUsernameQuery = `query { currentUser { id username } }`

type Config struct {
	URL   string   `json:"url"`
	Repos []string `json:"repos"`
}

type ExternalSvc struct {
	Kind        string `json:"Kind"`
	DisplayName string `json:"DisplayName"`
	Config      Config `json:"Config"`
}

type configWithToken struct {
	URL   string   `json:"url"`
	Repos []string `json:"repos"`
	Token string   `json:"token"`
}

var SourcegraphEndpoint = env.Get("SOURCEGRAPH_BASE_URL", "http://127.0.0.1:7080", "Sourcegraph frontend endpoint")
var SourcegraphAccessToken string

func main() {
	logfuncs := log.Init(log.Resource{
		Name: "executors test runner",
	})
	defer logfuncs.Sync()

	logger := log.Scoped("init", "runner initialization process")

	var client *gqltestutil.Client
	client, SourcegraphAccessToken = createSudoToken()

	f, err := os.Open("config/repos.json")
	if err != nil {
		logger.Fatal("Failed to open config/repos.json:", log.Error(err))
	}

	svcs := []ExternalSvc{}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&svcs); err != nil {
		f.Close()
		logger.Fatal("cannot parse repos.json", log.Error(err))
	}
	f.Close()

	if err := client.SetFeatureFlag("native-ssbc-execution", true); err != nil {
		logger.Fatal("Failed to set native-ssbc-execution feature flag", log.Error(err))
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	for _, svc := range svcs {
		b, _ := json.Marshal(configWithToken{
			Repos: svc.Config.Repos,
			URL:   svc.Config.URL,
			Token: githubToken,
		})

		_, err := client.AddExternalService(gqltestutil.AddExternalServiceInput{
			Kind:        svc.Kind,
			DisplayName: svc.DisplayName,
			Config:      string(b),
		})
		if err != nil {
			logger.Fatal("failed to add external service", log.String("name", svc.DisplayName), log.Error(err))
		}

		repos := []string{}
		for _, repo := range svc.Config.Repos {
			repos = append(repos, fmt.Sprintf("github.com/%s", repo))
		}

		logger.Info("waiting for repos to be cloned", log.Strings("repos", repos))

		if err = client.WaitForReposToBeCloned(repos...); err != nil {
			logger.Fatal("failed to wait for repos to be cloned", log.Error(err))
		}
	}

	// Now that we have our repositories synced and cloned into the instance, we
	// can start triggering an execution.
	batchSpecID, err := createBatchSpecForExecution(logger, client)
	if err != nil {
		logger.Fatal("failed to create batch spec for execution", log.Error(err))
	}

	// Now an execution has been enqueued. We wait for it to complete now.
	if err := awaitBatchSpecExecution(logger, client, batchSpecID); err != nil {
		logger.Fatal("failed to await batch spec execution", log.Error(err))
	}

	// Finally, we assert that the execution is in the correct shape.
	if err := assertBatchSpecExecution(logger, client, batchSpecID); err != nil {
		logger.Fatal("failed to assert batch spec execution", log.Error(err))
	}
}

const batchSpec = `
name: e2e-test-batch-change
description: Add Hello World to READMEs

on:
  - repository: github.com/sourcegraph/automation-testing
    branch: executors-e2e

steps:
  - run: IFS=$'\n'; echo Hello World | tee -a $(find -name README.md)
    container: ubuntu:18.04

changesetTemplate:
  title: Hello World
  body: My first batch change!
  branch: hello-world # Push the commit to this branch.
  commit:
    message: Append Hello World to all README.md files
`

func createBatchSpecForExecution(logger log.Logger, client *gqltestutil.Client) (string, error) {
	logger.Info("fetching user ID")

	id, err := client.CurrentUserID("")
	if err != nil {
		return "", err
	}

	logger.Info("Creating empty batch change")

	batchChangeID, err := client.CreateEmptyBatchChange(id, "e2e-test-batch-change")
	if err != nil {
		return "", err
	}

	logger.Info("Creating batch spec")

	batchSpecID, err := client.CreateBatchSpecFromRaw(batchChangeID, id, batchSpec)
	if err != nil {
		return "", err
	}

	logger.Info("Waiting for batch spec workspace resolution to finish")

	start := time.Now()
	for {
		if time.Now().Sub(start) > 60*time.Second {
			logger.Fatal("Waiting for batch spec workspace resolution to complete timed out")
		}
		state, err := client.GetBatchSpecWorkspaceResolutionStatus(batchSpecID)
		if err != nil {
			return "", err
		}
		if state == "COMPLETED" {
			break
		}

		if state == "FAILED" || state == "ERRORED" {
			logger.Fatal("Batch spec workspace resolution failed")
		}
	}

	logger.Info("Submitting execution for batch spec")

	// Execute with cache disabled.
	return batchSpecID, client.ExecuteBatchSpec(batchSpecID, true)
}

func awaitBatchSpecExecution(logger log.Logger, client *gqltestutil.Client, batchSpecID string) error {
	logger.Info("Waiting for batch spec execution to finish")

	start := time.Now()
	for {
		// Wait for at most 3 minutes.
		if time.Now().Sub(start) > 3*60*time.Second {
			logger.Fatal("Waiting for batch spec execution to complete timed out")
		}
		state, failureMessage, err := client.GetBatchSpecState(batchSpecID)
		if err != nil {
			return err
		}
		if state == "FAILED" {
			spec, err := client.GetBatchSpecDeep(batchSpecID)
			if err != nil {
				return err
			}
			d, err := json.MarshalIndent(spec, "", "")
			if err != nil {
				return err
			}
			logger.Fatal("Batch spec ended in failed state", log.String("failureMessage", failureMessage), log.String("spec", string(d)))
		}
		if state == "COMPLETED" {
			break
		}
	}
	return nil
}

func assertBatchSpecExecution(logger log.Logger, client *gqltestutil.Client, batchSpecID string) error {
	logger.Info("Loading batch spec to assert")

	batchSpec, err := client.GetBatchSpecDeep(batchSpecID)
	if err != nil {
		return err
	}
	if batchSpec.State != "COMPLETED" {
		logger.Fatal("batch spec is not complete")
	}
	if diff := cmp.Diff(batchSpec, &gqltestutil.BatchSpecDeep{
		ID:    batchSpecID,
		State: "COMPLETED",
		ChangesetSpecs: gqltestutil.BatchSpecChangesetSpecs{
			TotalCount: 1,
			Nodes: []gqltestutil.ChangesetSpec{
				{
					ID:   batchSpec.ChangesetSpecs.Nodes[0].ID,
					Type: "BRANCH",
					Description: gqltestutil.ChangesetSpecDescription{
						BaseRepository: gqltestutil.ChangesetRepository{Name: "github.com/sourcegraph/automation-testing"},
						BaseRef:        "executors-e2e",
						BaseRev:        "1c94aaf85d51e9d016b8ce4639b9f022d94c52e6",
						HeadRef:        "hello-world",
						Title:          "Hello World",
						Body:           "My first batch change!",
						Commits: []gqltestutil.ChangesetSpecCommit{
							{
								Message: "Append Hello World to all README.md files",
								Subject: "Append Hello World to all README.md files",
								Body:    "",
								Author: gqltestutil.ChangesetSpecCommitAuthor{
									Name:  "Sourcegraph",
									Email: "batch-changes@sourcegraph.com",
								},
							},
						},
						Diffs: gqltestutil.ChangesetSpecDiffs{
							FileDiffs: gqltestutil.ChangesetSpecFileDiffs{
								RawDiff: ``,
							},
						},
					},
					ForkTarget: gqltestutil.ChangesetForkTarget{},
				},
			},
		},
		CreatedAt:  batchSpec.CreatedAt,
		StartedAt:  batchSpec.StartedAt,
		FinishedAt: batchSpec.FinishedAt,
		Namespace:  gqltestutil.Namespace{ID: batchSpec.Namespace.ID},
		WorkspaceResolution: gqltestutil.WorkspaceResolution{
			Workspaces: gqltestutil.WorkspaceResolutionWorkspaces{
				TotalCount: 1,
				Stats: gqltestutil.WorkspaceResolutionWorkspacesStats{
					Completed: 1,
				},
				Nodes: []gqltestutil.BatchSpecWorkspace{
					{
						QueuedAt:   batchSpec.WorkspaceResolution.Workspaces.Nodes[0].QueuedAt,
						StartedAt:  batchSpec.WorkspaceResolution.Workspaces.Nodes[0].StartedAt,
						FinishedAt: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].FinishedAt,
						State:      "COMPLETED",
						DiffStat: gqltestutil.DiffStat{
							Added:   5,
							Deleted: 5,
						},
						Repository: gqltestutil.ChangesetRepository{Name: "github.com/sourcegraph/automation-testing"},
						Branch: gqltestutil.WorkspaceBranch{
							Name: "executors-e2e",
						},
						ChangesetSpecs: []gqltestutil.WorkspaceChangesetSpec{
							{
								ID: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].ChangesetSpecs[0].ID,
							},
						},
						SearchResultPaths: []string{},
						Executor: gqltestutil.Executor{
							Hostname:  batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Executor.Hostname,
							QueueName: "batches",
							Active:    true,
						},
						Stages: gqltestutil.BatchSpecWorkspaceStages{
							Setup: []gqltestutil.ExecutionLogEntry{
								{
									Key:     "setup.git.init",
									Command: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[0].Command,
									// Command: []string{
									// 	"git",
									// 	"-C",
									// 	"/scratch/workspace-1-4132878927/repository",
									// 	"init",
									// },
									StartTime: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[0].StartTime,
									ExitCode:  0,
									Out:       batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[0].Out,
									// "stderr: hint: Using 'master' as the name for the initial branch. This default branch name\nstderr: hint: is subject to change. To configure the initial branch name to use in all\nstderr: hint: of your new repositories, which will suppress this warning, call:\nstderr: hint: \nstderr: hint: \tgit config --global init.defaultBranch \u003cname\u003e\nstderr: hint: \nstderr: hint: Names commonly chosen instead of 'master' are 'main', 'trunk' and\nstderr: hint: 'development'. The just-created branch can be renamed via this command:\nstderr: hint: \nstderr: hint: \tgit branch -m \u003cname\u003e\nstdout: Initialized empty Git repository in /scratch/workspace-1-4132878927/repository/.git/\n",
									DurationMilliseconds: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[0].DurationMilliseconds,
								},
								{
									Key:     "setup.git.add-remote",
									Command: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[1].Command,
									// Command: []string{
									// 	"git",
									// 	"-C",
									// 	"/scratch/workspace-1-4132878927/repository",
									// 	"remote",
									// 	"add",
									// 	"origin",
									// 	"http://127.0.0.1:33123/github.com/sourcegraph/automation-testing",
									// },
									StartTime:            batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[1].StartTime,
									ExitCode:             0,
									Out:                  batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[1].Out,
									DurationMilliseconds: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[1].DurationMilliseconds,
								},
								{
									Key:     "setup.git.disable-gc",
									Command: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[2].Command,
									// Command: []string{
									// 	"git",
									// 	"-C",
									// 	"/scratch/workspace-1-4132878927/repository",
									// 	"config",
									// 	"--local",
									// 	"gc.auto",
									// 	"0",
									// },
									StartTime:            batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[2].StartTime,
									ExitCode:             0,
									Out:                  batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[2].Out,
									DurationMilliseconds: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[2].DurationMilliseconds,
								},
								{
									Key:     "setup.git.fetch",
									Command: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[3].Command,
									// Command: []string{
									// 	"git",
									// 	"-C",
									// 	"/scratch/workspace-1-4132878927/repository",
									// 	"-c",
									// 	"protocol.version=2",
									// 	"fetch",
									// 	"--progress",
									// 	"--no-recurse-submodules",
									// 	"--no-tags",
									// 	"--depth=1",
									// 	"origin",
									// 	"1c94aaf85d51e9d016b8ce4639b9f022d94c52e6",
									// },
									StartTime:            batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[3].StartTime,
									ExitCode:             0,
									Out:                  batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[3].Out,
									DurationMilliseconds: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[3].DurationMilliseconds,
									// "stderr: remote: Enumerating objects: 43, done.        \nstderr: remote: Counting objects:   2% (1/43)        \rremote: Counting objects:   4% (2/43)        \rremote: Counting objects:   6% (3/43)        \rremote: Counting objects:   9% (4/43)        \rremote: Counting objects:  11% (5/43)        \rremote: Counting objects:  13% (6/43)        \rremote: Counting objects:  16% (7/43)        \rremote: Counting objects:  18% (8/43)        \rremote: Counting objects:  20% (9/43)        \rremote: Counting objects:  23% (10/43)        \rremote: Counting objects:  25% (11/43)        \rremote: Counting objects:  27% (12/43)        \rremote: Counting objects:  30% (13/43)        \rremote: Counting objects:  32% (14/43)        \rremote: Counting objects:  34% (15/43)        \rremote: Counting objects:  37% (16/43)        \rremote: Counting objects:  39% (17/43)        \rremote: Counting objects:  41% (18/43)        \rremote: Counting objects:  44% (19/43)        \rremote: Counting objects:  46% (20/43)        \rremote: Counting objects:  48% (21/43)        \rremote: Counting objects:  51% (22/43)        \rremote: Counting objects:  53% (23/43)        \rremote: Counting objects:  55% (24/43)        \rremote: Counting objects:  58% (25/43)        \rremote: Counting objects:  60% (26/43)        \rremote: Counting objects:  62% (27/43)        \rremote: Counting objects:  65% (28/43)        \rremote: Counting objects:  67% (29/43)        \rremote: Counting objects:  69% (30/43)        \rremote: Counting objects:  72% (31/43)        \rremote: Counting objects:  74% (32/43)        \rremote: Counting objects:  76% (33/43)        \rremote: Counting objects:  79% (34/43)        \rremote: Counting objects:  81% (35/43)        \rremote: Counting objects:  83% (36/43)        \rremote: Counting objects:  86% (37/43)        \rremote: Counting objects:  88% (38/43)        \rremote: Counting objects:  90% (39/43)        \rremote: Counting objects:  93% (40/43)        \rremote: Counting objects:  95% (41/43)        \rremote: Counting objects:  97% (42/43)        \rremote: Counting objects: 100% (43/43)        \rremote: Counting objects: 100% (43/43), done.        \nstderr: remote: Compressing objects:   3% (1/32)        \rremote: Compressing objects:   6% (2/32)        \rremote: Compressing objects:   9% (3/32)        \rremote: Compressing objects:  12% (4/32)        \rremote: Compressing objects:  15% (5/32)        \rremote: Compressing objects:  18% (6/32)        \rremote: Compressing objects:  21% (7/32)        \rremote: Compressing objects:  25% (8/32)        \rremote: Compressing objects:  28% (9/32)        \rremote: Compressing objects:  31% (10/32)        \rremote: Compressing objects:  34% (11/32)        \rremote: Compressing objects:  37% (12/32)        \rremote: Compressing objects:  40% (13/32)        \rremote: Compressing objects:  43% (14/32)        \rremote: Compressing objects:  46% (15/32)        \rremote: Compressing objects:  50% (16/32)        \rremote: Compressing objects:  53% (17/32)        \rremote: Compressing objects:  56% (18/32)        \rremote: Compressing objects:  59% (19/32)        \rremote: Compressing objects:  62% (20/32)        \rremote: Compressing objects:  65% (21/32)        \rremote: Compressing objects:  68% (22/32)        \rremote: Compressing objects:  71% (23/32)        \rremote: Compressing objects:  75% (24/32)        \rremote: Compressing objects:  78% (25/32)        \rremote: Compressing objects:  81% (26/32)        \rremote: Compressing objects:  84% (27/32)        \rremote: Compressing objects:  87% (28/32)        \rremote: Compressing objects:  90% (29/32)        \rremote: Compressing objects:  93% (30/32)        \rremote: Compressing objects:  96% (31/32)        \rremote: Compressing objects: 100% (32/32)        \rremote: Compressing objects: 100% (32/32), done.        \nstderr: remote: Total 43 (delta 0), reused 38 (delta 0), pack-reused 0        \nstderr: From http://127.0.0.1:33123/github.com/sourcegraph/automation-testing\nstderr:  * branch            1c94aaf85d51e9d016b8ce4639b9f022d94c52e6 -\u003e FETCH_HEAD\n",
								},
								{
									Key:     "setup.git.checkout",
									Command: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[4].Command,
									// Command: []string{
									// 	"git",
									// 	"-C",
									// 	"/scratch/workspace-1-4132878927/repository",
									// 	"checkout",
									// 	"--progress",
									// 	"--force",
									// 	"1c94aaf85d51e9d016b8ce4639b9f022d94c52e6",
									// },
									StartTime:            batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[4].StartTime,
									ExitCode:             0,
									Out:                  batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[4].Out,
									DurationMilliseconds: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[4].DurationMilliseconds,
									// "stderr: Note: switching to '1c94aaf85d51e9d016b8ce4639b9f022d94c52e6'.\nstderr: \nstderr: You are in 'detached HEAD' state. You can look around, make experimental\nstderr: changes and commit them, and you can discard any commits you make in this\nstderr: state without impacting any branches by switching back to a branch.\nstderr: \nstderr: If you want to create a new branch to retain commits you create, you may\nstderr: do so (now or later) by using -c with the switch command. Example:\nstderr: \nstderr:   git switch -c \u003cnew-branch-name\u003e\nstderr: \nstderr: Or undo this operation with:\nstderr: \nstderr:   git switch -\nstderr: \nstderr: Turn off this advice by setting config variable advice.detachedHead to false\nstderr: \nstderr: HEAD is now at 1c94aaf Merge pull request #476 from sourcegraph/update-license\n",
								},
								{
									Key:     "setup.git.set-remote",
									Command: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[5].Command,
									// Command: []string{
									// 	"git",
									// 	"-C",
									// 	"/scratch/workspace-1-4132878927/repository",
									// 	"remote",
									// 	"set-url",
									// 	"origin",
									// 	"github.com/sourcegraph/automation-testing",
									// },
									StartTime:            batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[5].StartTime,
									ExitCode:             0,
									Out:                  batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[5].Out,
									DurationMilliseconds: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[5].DurationMilliseconds,
								},
								{
									Key:                  "setup.fs.extras",
									Command:              []string{},
									StartTime:            batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[6].StartTime,
									ExitCode:             0,
									Out:                  batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[6].Out,
									DurationMilliseconds: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Setup[6].DurationMilliseconds,
									// "Wrote /scratch/workspace-1-4132878927/.sourcegraph-executor/1.1_github.com_sourcegraph_automation-testing@1c94aaf85d51e9d016b8ce4639b9f022d94c52e6.sh in 1.305792ms\nWrote /scratch/workspace-1-4132878927/.sourcegraph-executor/1.2_github.com_sourcegraph_automation-testing@1c94aaf85d51e9d016b8ce4639b9f022d94c52e6.sh in 828.75µs\nWrote /scratch/workspace-1-4132878927/input.json in 931.125µs\nWrote /scratch/workspace-1-4132878927/.sourcegraph-executor/1.0_github.com_sourcegraph_automation-testing@1c94aaf85d51e9d016b8ce4639b9f022d94c52e6.sh in 777µs\n",
								},
							},
							SrcExec: []gqltestutil.ExecutionLogEntry{
								{
									Key:     "step.docker.step.0.pre",
									Command: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.SrcExec[0].Command,
									// Command: []string{
									// 	"docker",
									// 	"run",
									// 	"--rm",
									// 	"--cpus",
									// 	"4",
									// 	"--memory",
									// 	"12G",
									// 	"-v",
									// 	"/var/folders/q9/d_tcrm1j7gzfnt4f0sl9sygc0000gn/T/tmp.u6Sa7jhJ/executor-tmp/workspace-1-4132878927:/data",
									// 	"-w",
									// 	"/data",
									// 	"--entrypoint",
									// 	"/bin/sh",
									// 	"sourcegraph/batcheshelper:185194_2022-11-24_ca2c50198f68",
									// 	"/data/.sourcegraph-executor/1.0_github.com_sourcegraph_automation-testing@1c94aaf85d51e9d016b8ce4639b9f022d94c52e6.sh",
									// },
									StartTime:            batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.SrcExec[0].StartTime,
									ExitCode:             0,
									Out:                  batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.SrcExec[0].Out,
									DurationMilliseconds: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.SrcExec[0].DurationMilliseconds,
									// "stderr: WARNING: The requested image's platform (linux/amd64) does not match the detected host platform (linux/arm64/v8) and no specific platform was requested\nstderr: + batcheshelper pre 0\n",
								},
								{
									Key:     "step.docker.step.0.run",
									Command: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.SrcExec[1].Command,
									// Command: []string{
									// 	"docker",
									// 	"run",
									// 	"--rm",
									// 	"--cpus",
									// 	"4",
									// 	"--memory",
									// 	"12G",
									// 	"-v",
									// 	"/var/folders/q9/d_tcrm1j7gzfnt4f0sl9sygc0000gn/T/tmp.u6Sa7jhJ/executor-tmp/workspace-1-4132878927:/data",
									// 	"-w",
									// 	"/data/repository",
									// 	"--entrypoint",
									// 	"/bin/sh",
									// 	"ubuntu:18.04",
									// 	"/data/.sourcegraph-executor/1.1_github.com_sourcegraph_automation-testing@1c94aaf85d51e9d016b8ce4639b9f022d94c52e6.sh",
									// },
									StartTime:            batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.SrcExec[1].StartTime,
									ExitCode:             0,
									Out:                  batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.SrcExec[1].Out,
									DurationMilliseconds: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.SrcExec[1].DurationMilliseconds,
									// "stderr: WARNING: The requested image's platform (linux/amd64) does not match the detected host platform (linux/arm64/v8) and no specific platform was requested\nstderr: Hello World\n",
								},
								{
									Key:     "step.docker.step.0.post",
									Command: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.SrcExec[2].Command,
									// Command: []string{
									// 	"docker",
									// 	"run",
									// 	"--rm",
									// 	"--cpus",
									// 	"4",
									// 	"--memory",
									// 	"12G",
									// 	"-v",
									// 	"/var/folders/q9/d_tcrm1j7gzfnt4f0sl9sygc0000gn/T/tmp.u6Sa7jhJ/executor-tmp/workspace-1-4132878927:/data",
									// 	"-w",
									// 	"/data",
									// 	"--entrypoint",
									// 	"/bin/sh",
									// 	"sourcegraph/batcheshelper:185194_2022-11-24_ca2c50198f68",
									// 	"/data/.sourcegraph-executor/1.2_github.com_sourcegraph_automation-testing@1c94aaf85d51e9d016b8ce4639b9f022d94c52e6.sh",
									// },
									StartTime:            batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.SrcExec[2].StartTime,
									ExitCode:             0,
									Out:                  batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.SrcExec[2].Out,
									DurationMilliseconds: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.SrcExec[2].DurationMilliseconds,
									// "stderr: WARNING: The requested image's platform (linux/amd64) does not match the detected host platform (linux/arm64/v8) and no specific platform was requested\nstderr: + batcheshelper post 0\nstdout: {\"operation\":\"CACHE_AFTER_STEP_RESULT\",\"timestamp\":\"2022-11-24T22:56:22.696Z\",\"status\":\"SUCCESS\",\"metadata\":{\"key\":\"wHcoEItqNkdJj9k1-1sRCQ-step-0\",\"value\":{\"changedFiles\":{\"modified\":null,\"added\":null,\"deleted\":null,\"renamed\":null},\"stdout\":\"Hello World\\n\",\"stderr\":\"\",\"stepIndex\":0,\"diff\":\"diff --git README.md README.md\\nindex 1914491..89e55af 100644\\n--- README.md\\n+++ README.md\\n@@ -3,4 +3,4 @@ This repository is used to test opening and closing pull request with Automation\\n \\n (c) Copyright Sourcegraph 2013-2020.\\n (c) Copyright Sourcegraph 2013-2020.\\n-(c) Copyright Sourcegraph 2013-2020.\\n\\\\ No newline at end of file\\n+(c) Copyright Sourcegraph 2013-2020.Hello World\\ndiff --git examples/README.md examples/README.md\\nindex 40452a9..a32cc2f 100644\\n--- examples/README.md\\n+++ examples/README.md\\n@@ -5,4 +5,4 @@ This folder contains examples\\n (This is a test message, ignore)\\n \\n (c) Copyright Sourcegraph 2013-2020.\\n-(c) Copyright Sourcegraph 2013-2020.\\n\\\\ No newline at end of file\\n+(c) Copyright Sourcegraph 2013-2020.Hello World\\ndiff --git examples/project3/README.md examples/project3/README.md\\nindex 272d9ea..f49f17d 100644\\n--- examples/project3/README.md\\n+++ examples/project3/README.md\\n@@ -1,4 +1,4 @@\\n # project3\\n \\n (c) Copyright Sourcegraph 2013-2020.\\n-(c) Copyright Sourcegraph 2013-2020.\\n\\\\ No newline at end of file\\n+(c) Copyright Sourcegraph 2013-2020.Hello World\\ndiff --git project1/README.md project1/README.md\\nindex 8a5f437..6284591 100644\\n--- project1/README.md\\n+++ project1/README.md\\n@@ -3,4 +3,4 @@\\n This is project 1.\\n \\n (c) Copyright Sourcegraph 2013-2020.\\n-(c) Copyright Sourcegraph 2013-2020.\\n\\\\ No newline at end of file\\n+(c) Copyright Sourcegraph 2013-2020.Hello World\\ndiff --git project2/README.md project2/README.md\\nindex b1e1cdd..9445efe 100644\\n--- project2/README.md\\n+++ project2/README.md\\n@@ -1,3 +1,3 @@\\n This is a starter template for [Learn Next.js](https://nextjs.org/learn).\\n (c) Copyright Sourcegraph 2013-2020.\\n-(c) Copyright Sourcegraph 2013-2020.\\n\\\\ No newline at end of file\\n+(c) Copyright Sourcegraph 2013-2020.Hello World\\n\",\"outputs\":{}}}}\n",
								},
							},
							Teardown: []gqltestutil.ExecutionLogEntry{
								{
									Key:                  "teardown.fs",
									Command:              []string{},
									StartTime:            batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Teardown[0].StartTime,
									ExitCode:             0,
									Out:                  batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Teardown[0].Out,
									DurationMilliseconds: batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Stages.Teardown[0].DurationMilliseconds,
									// "Removing /scratch/workspace-1-4132878927\n",
								},
							},
						},
						Steps: []gqltestutil.BatchSpecWorkspaceStep{
							{
								Number:    1,
								Run:       "IFS=$'\\n'; echo Hello World | tee -a $(find -name README.md)",
								Container: "ubuntu:18.04",
								OutputLines: []string{
									"stderr: WARNING: The requested image's platform (linux/amd64) does not match the detected host platform (linux/arm64/v8) and no specific platform was requested",
									"stderr: Hello World",
									"",
								},
								StartedAt:       batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Steps[0].StartedAt,
								FinishedAt:      batchSpec.WorkspaceResolution.Workspaces.Nodes[0].Steps[0].FinishedAt,
								ExitCode:        0,
								Environment:     []gqltestutil.WorkspaceEnvironmentVariable{},
								OutputVariables: []gqltestutil.WorkspaceOutputVariable{},
								DiffStat: gqltestutil.DiffStat{
									Added:   5,
									Deleted: 5,
								},
								Diff: gqltestutil.ChangesetSpecDiffs{
									FileDiffs: gqltestutil.ChangesetSpecFileDiffs{
										RawDiff: "diff --git README.md README.md\nindex 1914491..89e55af 100644\n--- README.md\n+++ README.md\n@@ -3,4 +3,4 @@ This repository is used to test opening and closing pull request with Automation\n \n (c) Copyright Sourcegraph 2013-2020.\n (c) Copyright Sourcegraph 2013-2020.\n-(c) Copyright Sourcegraph 2013-2020.\n\\ No newline at end of file\n+(c) Copyright Sourcegraph 2013-2020.Hello World\ndiff --git examples/README.md examples/README.md\nindex 40452a9..a32cc2f 100644\n--- examples/README.md\n+++ examples/README.md\n@@ -5,4 +5,4 @@ This folder contains examples\n (This is a test message, ignore)\n \n (c) Copyright Sourcegraph 2013-2020.\n-(c) Copyright Sourcegraph 2013-2020.\n\\ No newline at end of file\n+(c) Copyright Sourcegraph 2013-2020.Hello World\ndiff --git examples/project3/README.md examples/project3/README.md\nindex 272d9ea..f49f17d 100644\n--- examples/project3/README.md\n+++ examples/project3/README.md\n@@ -1,4 +1,4 @@\n # project3\n \n (c) Copyright Sourcegraph 2013-2020.\n-(c) Copyright Sourcegraph 2013-2020.\n\\ No newline at end of file\n+(c) Copyright Sourcegraph 2013-2020.Hello World\ndiff --git project1/README.md project1/README.md\nindex 8a5f437..6284591 100644\n--- project1/README.md\n+++ project1/README.md\n@@ -3,4 +3,4 @@\n This is project 1.\n \n (c) Copyright Sourcegraph 2013-2020.\n-(c) Copyright Sourcegraph 2013-2020.\n\\ No newline at end of file\n+(c) Copyright Sourcegraph 2013-2020.Hello World\ndiff --git project2/README.md project2/README.md\nindex b1e1cdd..9445efe 100644\n--- project2/README.md\n+++ project2/README.md\n@@ -1,3 +1,3 @@\n This is a starter template for [Learn Next.js](https://nextjs.org/learn).\n (c) Copyright Sourcegraph 2013-2020.\n-(c) Copyright Sourcegraph 2013-2020.\n\\ No newline at end of file\n+(c) Copyright Sourcegraph 2013-2020.Hello World\n",
									},
								},
							},
						},
					},
				},
			},
		},
		ExpiresAt: batchSpec.ExpiresAt,
		Source:    "REMOTE",
		Files: gqltestutil.BatchSpecFiles{
			TotalCount: 0,
			Nodes:      []gqltestutil.BatchSpecFile{},
		},
	}); diff != "" {
		fmt.Printf("Batch spec diff detected: %s\n", diff)
		return errors.New("batch spec not in expected state")
	}
	return nil
}
