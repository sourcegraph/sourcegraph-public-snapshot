package main

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
)

const (
	sourceRepository = "github.com/sourcegraph/automation-testing"
	sourceRef        = "executors-e2e"
	authorName       = "sourcegraph"
	authorEmail      = "sourcegraph@sourcegraph.com"
	changeSetBody    = "My first batch change!"
	container        = "alpine:3"
)

func testHelloWorldBatchChange() Test {
	batchSpecPs := batchSpecParams{
		NameContent:             "hello-world",
		Description:             "Add Hello World to READMEs",
		RunCommand:              "IFS=$'\\n'; echo Hello World | tee -a $(find -name README.md)",
		ChangeSetTemplateTitle:  "Hello World",
		ChangeSetTemplateBranch: "hello-world",
		CommitMessage:           "Append Hello World to all README.md files",
	}
	batchSpec := generateBatchSpec(batchSpecPs)

	diffPs := diffParams{
		READMEObjectHash:         "89e55af",
		ExamplesREADMEObjectHash: "a32cc2f",
		Project3READMEObjectHash: "f49f17d",
		Project1READMEObjectHash: "6284591",
		Project2READMEObjectHash: "9445efe",
		HelloWorldMessage:        "Hello World",
	}
	expectedDiff := generateDiff(diffPs)

	expectedState := gqltestutil.BatchSpecDeep{
		State: "COMPLETED",
		ChangesetSpecs: gqltestutil.BatchSpecChangesetSpecs{
			TotalCount: 1,
			Nodes: []gqltestutil.ChangesetSpec{
				{
					Type: "BRANCH",
					Description: gqltestutil.ChangesetSpecDescription{
						BaseRepository: gqltestutil.ChangesetRepository{Name: sourceRepository},
						BaseRef:        sourceRef,
						BaseRev:        "1c94aaf85d51e9d016b8ce4639b9f022d94c52e6",
						HeadRef:        batchSpecPs.ChangeSetTemplateBranch,
						Title:          batchSpecPs.ChangeSetTemplateTitle,
						Body:           changeSetBody,
						Commits: []gqltestutil.ChangesetSpecCommit{
							{
								Message: batchSpecPs.CommitMessage,
								Subject: batchSpecPs.CommitMessage,
								Body:    "",
								Author: gqltestutil.ChangesetSpecCommitAuthor{
									Name:  authorName,
									Email: authorEmail,
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
		Namespace: gqltestutil.Namespace{},
		WorkspaceResolution: gqltestutil.WorkspaceResolution{
			Workspaces: gqltestutil.WorkspaceResolutionWorkspaces{
				TotalCount: 1,
				Stats: gqltestutil.WorkspaceResolutionWorkspacesStats{
					Completed: 1,
				},
				Nodes: []gqltestutil.BatchSpecWorkspace{
					{
						State: "COMPLETED",
						DiffStat: gqltestutil.DiffStat{
							Added:   5,
							Deleted: 5,
						},
						Repository: gqltestutil.ChangesetRepository{Name: sourceRepository},
						Branch: gqltestutil.WorkspaceBranch{
							Name: sourceRef,
						},
						ChangesetSpecs: []gqltestutil.WorkspaceChangesetSpec{
							{},
						},
						SearchResultPaths: []string{},
						Executor: gqltestutil.Executor{
							QueueName: "batches",
							Active:    true,
						},
						Stages: gqltestutil.BatchSpecWorkspaceStages{
							Setup: []gqltestutil.ExecutionLogEntry{
								{
									Key:      "setup.git.init",
									ExitCode: 0,
								},
								{
									Key:      "setup.git.add-remote",
									ExitCode: 0,
								},
								{
									Key:      "setup.git.disable-gc",
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
									Key:      "setup.fs.extras",
									Command:  []string{},
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
							Teardown: []gqltestutil.ExecutionLogEntry{
								{
									Key:      "teardown.fs",
									Command:  []string{},
									ExitCode: 0,
								},
							},
						},
						Steps: []gqltestutil.BatchSpecWorkspaceStep{
							{
								Number:    1,
								Run:       batchSpecPs.RunCommand,
								Container: container,
								OutputLines: gqltestutil.WorkspaceOutputLines{
									Nodes: []string{
										"stderr: WARNING: The requested image's platform (linux/amd64) does not match the detected host platform (linux/arm64/v8) and no specific platform was requested",
										"stderr: Hello World",
										"",
									},
									TotalCount: 3,
								},
								ExitCode:        0,
								Environment:     []gqltestutil.WorkspaceEnvironmentVariable{},
								OutputVariables: []gqltestutil.WorkspaceOutputVariable{},
								DiffStat: gqltestutil.DiffStat{
									Added:   5,
									Deleted: 5,
								},
								Diff: gqltestutil.ChangesetSpecDiffs{
									FileDiffs: gqltestutil.ChangesetSpecFileDiffs{
										RawDiff: expectedDiff,
									},
								},
							},
						},
					},
				},
			},
		},
		Source: "REMOTE",
		Files: gqltestutil.BatchSpecFiles{
			TotalCount: 0,
			Nodes:      []gqltestutil.BatchSpecFile{},
		},
	}

	return Test{
		PreExistingCacheEntries: map[string]execution.AfterStepResult{},
		BatchSpecInput:          batchSpec,
		ExpectedCacheEntries: map[string]execution.AfterStepResult{
			"d4ndKwesInT_CAoLz8351A-step-0": {
				Version: 2,
				Stdout:  fmt.Sprintf("%s\n", diffPs.HelloWorldMessage),
				Diff:    []byte(expectedDiff),
				Outputs: map[string]any{},
			},
		},
		ExpectedChangesetSpecs: []*types.ChangesetSpec{
			{
				Type:              "branch",
				DiffStatAdded:     5,
				DiffStatDeleted:   5,
				BatchSpecID:       2,
				BaseRepoID:        1,
				UserID:            1,
				BaseRev:           "1c94aaf85d51e9d016b8ce4639b9f022d94c52e6",
				BaseRef:           sourceRef,
				HeadRef:           fmt.Sprintf("refs/heads/%s", batchSpecPs.ChangeSetTemplateBranch),
				Title:             batchSpecPs.ChangeSetTemplateTitle,
				Body:              changeSetBody,
				CommitMessage:     batchSpecPs.CommitMessage,
				CommitAuthorName:  authorName,
				CommitAuthorEmail: authorEmail,
				Diff:              []byte(expectedDiff),
			},
		},
		ExpectedState: expectedState,
		CacheDisabled: true,
	}
}

func testEnvObjectBatchChange() Test {
	batchSpecPs := batchSpecParams{
		NameContent: "env-object",
		Description: "Add the value of an environment variable object to READMEs",
		RunCommand:  "IFS=$'\\n'; echo $MESSAGE | tee -a $(find -name README.md)",
		AdditionalKeyValues: []specStepBlock{
			{
				BlockName: "env",
				KeyValues: []keyValue{
					{Key: "MESSAGE", Value: "Hello world from an env object!"},
				},
			},
		},
		ChangeSetTemplateTitle:  "Hello World from env object",
		ChangeSetTemplateBranch: "env-object",
		CommitMessage:           "Append an env object to all README.md files",
	}
	batchSpec := generateBatchSpec(batchSpecPs)

	diffPs := diffParams{
		READMEObjectHash:         "23aa51b",
		ExamplesREADMEObjectHash: "3705d13",
		Project3READMEObjectHash: "140c423",
		Project1READMEObjectHash: "3075ce8",
		Project2READMEObjectHash: "0fb42ff",
		HelloWorldMessage:        "Hello world from an env object!",
	}
	expectedDiff := generateDiff(diffPs)

	expectedState := gqltestutil.BatchSpecDeep{
		State: "COMPLETED",
		ChangesetSpecs: gqltestutil.BatchSpecChangesetSpecs{
			TotalCount: 1,
			Nodes: []gqltestutil.ChangesetSpec{
				{
					Type: "BRANCH",
					Description: gqltestutil.ChangesetSpecDescription{
						BaseRepository: gqltestutil.ChangesetRepository{Name: sourceRepository},
						BaseRef:        sourceRef,
						BaseRev:        "1c94aaf85d51e9d016b8ce4639b9f022d94c52e6",
						HeadRef:        batchSpecPs.ChangeSetTemplateBranch,
						Title:          batchSpecPs.ChangeSetTemplateTitle,
						Body:           changeSetBody,
						Commits: []gqltestutil.ChangesetSpecCommit{
							{
								Message: batchSpecPs.CommitMessage,
								Subject: batchSpecPs.CommitMessage,
								Body:    "",
								Author: gqltestutil.ChangesetSpecCommitAuthor{
									Name:  authorName,
									Email: authorEmail,
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
		Namespace: gqltestutil.Namespace{},
		WorkspaceResolution: gqltestutil.WorkspaceResolution{
			Workspaces: gqltestutil.WorkspaceResolutionWorkspaces{
				TotalCount: 1,
				Stats: gqltestutil.WorkspaceResolutionWorkspacesStats{
					Completed: 1,
				},
				Nodes: []gqltestutil.BatchSpecWorkspace{
					{
						State: "COMPLETED",
						DiffStat: gqltestutil.DiffStat{
							Added:   5,
							Deleted: 5,
						},
						Repository: gqltestutil.ChangesetRepository{Name: sourceRepository},
						Branch: gqltestutil.WorkspaceBranch{
							Name: sourceRef,
						},
						ChangesetSpecs: []gqltestutil.WorkspaceChangesetSpec{
							{},
						},
						SearchResultPaths: []string{},
						Executor: gqltestutil.Executor{
							QueueName: "batches",
							Active:    true,
						},
						Stages: gqltestutil.BatchSpecWorkspaceStages{
							Setup: []gqltestutil.ExecutionLogEntry{
								{
									Key:      "setup.git.init",
									ExitCode: 0,
								},
								{
									Key:      "setup.git.add-remote",
									ExitCode: 0,
								},
								{
									Key:      "setup.git.disable-gc",
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
									Key:      "setup.fs.extras",
									Command:  []string{},
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
							Teardown: []gqltestutil.ExecutionLogEntry{
								{
									Key:      "teardown.fs",
									Command:  []string{},
									ExitCode: 0,
								},
							},
						},
						Steps: []gqltestutil.BatchSpecWorkspaceStep{
							{
								Number:    1,
								Run:       batchSpecPs.RunCommand,
								Container: container,
								OutputLines: gqltestutil.WorkspaceOutputLines{
									Nodes: []string{
										"stderr: WARNING: The requested image's platform (linux/amd64) does not match the detected host platform (linux/arm64/v8) and no specific platform was requested",
										"stderr: Hello World",
										"",
									},
									TotalCount: 3,
								},
								ExitCode: 0,
								Environment: []gqltestutil.WorkspaceEnvironmentVariable{
									{
										Name:  batchSpecPs.AdditionalKeyValues[0].KeyValues[0].Key,
										Value: batchSpecPs.AdditionalKeyValues[0].KeyValues[0].Value,
									},
								},
								OutputVariables: []gqltestutil.WorkspaceOutputVariable{},
								DiffStat: gqltestutil.DiffStat{
									Added:   5,
									Deleted: 5,
								},
								Diff: gqltestutil.ChangesetSpecDiffs{
									FileDiffs: gqltestutil.ChangesetSpecFileDiffs{
										RawDiff: expectedDiff,
									},
								},
							},
						},
					},
				},
			},
		},
		Source: "REMOTE",
		Files: gqltestutil.BatchSpecFiles{
			TotalCount: 0,
			Nodes:      []gqltestutil.BatchSpecFile{},
		},
	}

	return Test{
		PreExistingCacheEntries: map[string]execution.AfterStepResult{},
		BatchSpecInput:          batchSpec,
		ExpectedCacheEntries: map[string]execution.AfterStepResult{
			"IZ_d2HAMbc9BDhI2uWpavA-step-0": {
				Version: 2,
				Stdout:  fmt.Sprintf("%s\n", diffPs.HelloWorldMessage),
				Diff:    []byte(expectedDiff),
				Outputs: map[string]any{},
			},
		},
		ExpectedChangesetSpecs: []*types.ChangesetSpec{
			{
				Type:              "branch",
				DiffStatAdded:     5,
				DiffStatDeleted:   5,
				BatchSpecID:       2,
				BaseRepoID:        1,
				UserID:            1,
				BaseRev:           "1c94aaf85d51e9d016b8ce4639b9f022d94c52e6",
				BaseRef:           sourceRef,
				HeadRef:           fmt.Sprintf("refs/heads/%s", batchSpecPs.ChangeSetTemplateBranch),
				Title:             batchSpecPs.ChangeSetTemplateTitle,
				Body:              changeSetBody,
				CommitMessage:     batchSpecPs.CommitMessage,
				CommitAuthorName:  authorName,
				CommitAuthorEmail: authorEmail,
				Diff:              []byte(expectedDiff),
			},
		},
		ExpectedState: expectedState,
		CacheDisabled: true,
	}
}

func testFileMountBatchChange() Test {

	return Test{}
}

type diffParams struct {
	READMEObjectHash         string
	ExamplesREADMEObjectHash string
	Project3READMEObjectHash string
	Project1READMEObjectHash string
	Project2READMEObjectHash string
	HelloWorldMessage        string
}

func generateDiff(params diffParams) string {
	const diffTemplateString = `diff --git README.md README.md
index 1914491..{{.READMEObjectHash}} 100644
--- README.md
+++ README.md
@@ -3,4 +3,4 @@ This repository is used to test opening and closing pull request with Automation
 ` + `
 (c) Copyright Sourcegraph 2013-2020.
 (c) Copyright Sourcegraph 2013-2020.
-(c) Copyright Sourcegraph 2013-2020.
\ No newline at end of file
+(c) Copyright Sourcegraph 2013-2020.{{.HelloWorldMessage}}
diff --git examples/README.md examples/README.md
index 40452a9..{{.ExamplesREADMEObjectHash}} 100644
--- examples/README.md
+++ examples/README.md
@@ -5,4 +5,4 @@ This folder contains examples
 (This is a test message, ignore)
 ` + `
 (c) Copyright Sourcegraph 2013-2020.
-(c) Copyright Sourcegraph 2013-2020.
\ No newline at end of file
+(c) Copyright Sourcegraph 2013-2020.{{.HelloWorldMessage}}
diff --git examples/project3/README.md examples/project3/README.md
index 272d9ea..{{.Project3READMEObjectHash}} 100644
--- examples/project3/README.md
+++ examples/project3/README.md
@@ -1,4 +1,4 @@
 # project3
 ` + `
 (c) Copyright Sourcegraph 2013-2020.
-(c) Copyright Sourcegraph 2013-2020.
\ No newline at end of file
+(c) Copyright Sourcegraph 2013-2020.{{.HelloWorldMessage}}
diff --git project1/README.md project1/README.md
index 8a5f437..{{.Project1READMEObjectHash}} 100644
--- project1/README.md
+++ project1/README.md
@@ -3,4 +3,4 @@
 This is project 1.
 ` + `
 (c) Copyright Sourcegraph 2013-2020.
-(c) Copyright Sourcegraph 2013-2020.
\ No newline at end of file
+(c) Copyright Sourcegraph 2013-2020.{{.HelloWorldMessage}}
diff --git project2/README.md project2/README.md
index b1e1cdd..{{.Project2READMEObjectHash}} 100644
--- project2/README.md
+++ project2/README.md
@@ -1,3 +1,3 @@
 This is a starter template for [Learn Next.js](https://nextjs.org/learn).
 (c) Copyright Sourcegraph 2013-2020.
-(c) Copyright Sourcegraph 2013-2020.
\ No newline at end of file
+(c) Copyright Sourcegraph 2013-2020.{{.HelloWorldMessage}}
`

	tmpl, err := template.New("diffTemplate").Parse(diffTemplateString)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, params); err != nil {
		panic(err)
	}

	return buf.String()
}

type keyValue struct {
	Key   string
	Value string
}

type specStepBlock struct {
	BlockName string
	KeyValues []keyValue
}

type batchSpecParams struct {
	NameContent             string
	Description             string
	RunCommand              string
	AdditionalKeyValues     []specStepBlock
	ChangeSetTemplateTitle  string
	ChangeSetTemplateBranch string
	CommitMessage           string
}

func generateBatchSpec(params batchSpecParams) string {
	batchSpecTemplateString := `
name: e2e-test-batch-change-{{.NameContent}}
description: {{.Description}}

on:
  - repository: github.com/sourcegraph/automation-testing
    # This branch is never changing - hopefully.
    branch: executors-e2e

steps:
  - run: {{.RunCommand}}
    container: alpine:3
{{range .AdditionalKeyValues}}
    {{.BlockName}}:
    {{range .KeyValues}}
      {{.Key}}: {{.Value}}
    {{end}}
{{end}}

changesetTemplate:
  title: {{.ChangeSetTemplateTitle}}
  body: My first batch change!
  branch: {{.ChangeSetTemplateBranch}} # Push the commit to this branch.
  commit:
    message: {{.CommitMessage}}
`

	tmpl, err := template.New("batchSpecTemplate").Parse(batchSpecTemplateString)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, params); err != nil {
		panic(err)
	}

	return buf.String()
}
