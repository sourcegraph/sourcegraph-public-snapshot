package campaigns

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/neelance/parallel"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/repos/git"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/diagnostics"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gituri"
)

type rulesExecutor struct {
	input graphqlbackend.CreateCampaignInput
}

func (x *rulesExecutor) planThreads(ctx context.Context) ([]graphqlbackend.ThreadPreview, error) {
	diagnostics, err := extdata{}.parseDiagnosticInfos(x.input.ExtensionData)
	if err != nil {
		return nil, err
	}
	diagnosticsByRepo := map[api.RepoID][]diagnosticInfo{}
	for _, d := range diagnostics {
		repo, err := backend.Repos.GetByName(ctx, d.ResourceURI.Repo())
		if err != nil {
			return nil, err
		}
		diagnosticsByRepo[repo.ID] = append(diagnosticsByRepo[repo.ID], d)
	}

	toRawDiagnostics := func(diags []diagnosticInfo) []string {
		rawDiagnostics := make([]string, len(diags))
		for i, d := range diags {
			b, err := json.Marshal(d.RawDiagnostic)
			if err != nil {
				panic(err)
			}
			rawDiagnostics[i] = string(b)
		}
		return rawDiagnostics
	}

	planIssues := func(ctx context.Context) (issues []graphqlbackend.ThreadPreview, err error) {
		// TODO!(sqs): hack, if there are any RawFileDiffs, assume the whole campaign is changesets
		// only and issues arent wanted.
		if len(x.input.ExtensionData.RawFileDiffs) > 0 {
			return nil, nil
		}

		// Include issues for each diagnostic.
		for repoID, diagnostics := range diagnosticsByRepo {
			repo, err := backend.Repos.Get(ctx, repoID)
			if err != nil {
				return nil, err
			}

			// Use the diagnostic message if all are the same; otherwise, use the first and mention
			// the others.
			var title string
			for _, d := range diagnostics {
				if title == "" {
					title = d.Message
				} else if title != d.Message {
					title = fmt.Sprintf("%s (+%d others)", title, len(diagnostics)-1)
					break
				}
			}

			rawDiagnostics := toRawDiagnostics(diagnostics)
			issues = append(issues, threads.NewGQLThreadPreview(graphqlbackend.CreateThreadInput{
				Repository:     graphqlbackend.NewRepositoryResolver(repo).ID(),
				Title:          title,
				Body:           x.input.Body,
				RawDiagnostics: &rawDiagnostics,
			}, nil))
		}
		return issues, nil
	}

	planChangesets := func(ctx context.Context) (changesets []graphqlbackend.ThreadPreview, err error) {
		// Include changesets for each diff.
		diffs, err := extdata{}.parseRawFileDiffs(x.input.ExtensionData)
		if err != nil {
			return nil, err
		}
		byRepo := map[api.RepoID]*git.GQLRepositoryComparisonPreview{}
		for _, d := range diffs {
			newNameURI, err := gituri.Parse(d.NewName)
			if err != nil {
				return nil, err
			}
			repo, err := backend.Repos.GetByName(ctx, newNameURI.Repo())
			if err != nil {
				return nil, err
			}
			repoComparison, ok := byRepo[repo.ID]
			if !ok {
				repoComparison = &git.GQLRepositoryComparisonPreview{
					BaseRepository_: graphqlbackend.NewRepositoryResolver(repo),
					HeadRepository_: graphqlbackend.NewRepositoryResolver(repo),
				}
				byRepo[repo.ID] = repoComparison
			}
			repoComparison.FileDiffs_ = append(repoComparison.FileDiffs_, d)
		}
		for repoID, repoComparison := range byRepo {
			repo, err := backend.Repos.Get(ctx, repoID)
			if err != nil {
				return nil, err
			}

			rawDiagnostics := toRawDiagnostics(diagnosticsByRepo[repoID])
			changesets = append(changesets, threads.NewGQLThreadPreview(graphqlbackend.CreateThreadInput{
				Repository:     graphqlbackend.NewRepositoryResolver(repo).ID(),
				Title:          x.input.Name,
				Body:           x.input.Body,
				RawDiagnostics: &rawDiagnostics,
			}, repoComparison))
		}
		return changesets, nil
	}
	// TODO!(sqs): include existing issues/threads matched by rules

	issues, err := planIssues(ctx)
	if err != nil {
		return nil, err
	}
	changesets, err := planChangesets(ctx)
	if err != nil {
		return nil, err
	}
	allThreads := append(issues, changesets...)
	sort.Slice(allThreads, func(i, j int) bool {
		return allThreads[i].Internal_RepositoryID() < allThreads[j].Internal_RepositoryID()
	})
	return allThreads, nil
}

func (x *rulesExecutor) syncThreads(ctx context.Context, campaignID int64, campaignName string) error {
	allThreads, err := x.planThreads(ctx)
	if err != nil {
		return err
	}

	// TODO!(sqs): sync issues too - right now we only sync changesets because they are easier to
	// sync because they have a base/head that uniquely identifies them.
	run := parallel.NewRun(16)
	for _, thread := range allThreads {
		run.Acquire()
		// TODO!(sqs): use goroutine.Go, but it's in cmd/frontend/internal
		go func(thread graphqlbackend.ThreadPreview) {
			defer run.Release()
			kind, err := thread.Kind(ctx)
			if err != nil {
				run.Error(err)
				return
			}
			switch kind {
			case graphqlbackend.ThreadKindChangeset:
				if err := x.syncChangeset(ctx, campaignID, campaignName, thread); err != nil {
					run.Error(err)
					return
				}
				// TODO!(sqs): support issues, discussions (other thread kinds)
			}
		}(thread)
	}
	return run.Wait()
}

func (x *rulesExecutor) syncChangeset(ctx context.Context, campaignID int64, campaignName string, thread graphqlbackend.ThreadPreview) error {
	repo, err := thread.Repository(ctx)
	if err != nil {
		return err
	}

	repoComparison, err := thread.RepositoryComparison(ctx)
	if err != nil {
		return err
	}
	fileDiffConnection := repoComparison.FileDiffs(&graphqlutil.ConnectionArgs{})
	if err != nil {
		return err
	}
	patch, err := fileDiffConnection.RawDiff(ctx)
	if err != nil {
		return err
	}
	// Convert full URIs in patch to relative URIs.
	fileDiffs, err := diff.ParseMultiFileDiff([]byte(patch))
	if err != nil {
		return err
	}
	for _, d := range fileDiffs {
		pathOnly := func(uriStr string) string {
			u, err := gituri.Parse(uriStr)
			if err != nil {
				panic(err)
			}
			return u.FilePath()
		}
		d.OrigName = "a/" + pathOnly(d.OrigName)
		d.NewName = "b/" + pathOnly(d.NewName)
	}
	patchBytes, err := diff.PrintMultiFileDiff(fileDiffs)
	if err != nil {
		return err
	}
	patch = string(patchBytes)

	defaultBranch, err := repo.DefaultBranch(ctx)
	if err != nil {
		return err
	}
	oid, err := defaultBranch.Target().OID(ctx)
	if err != nil {
		return err
	}
	var IsAlphanumericWithPeriod = regexp.MustCompile(`[^a-zA-Z0-9_.]+`)
	branchName := "a8n/" + strings.TrimSuffix(IsAlphanumericWithPeriod.ReplaceAllString(x.input.Name, "-"), "-") // TODO!(sqs): hack
	_, err = git.GraphQLResolver{}.CreateRefFromPatch(ctx, &struct {
		Input graphqlbackend.GitCreateRefFromPatchInput
	}{
		Input: graphqlbackend.GitCreateRefFromPatchInput{
			Repository:    repo.ID(),
			Name:          "refs/heads/" + branchName, //TODO!(sqs)
			BaseCommit:    oid,
			Patch:         patch,
			CommitMessage: "a8n: " + x.input.Name,
		},
	})
	if err != nil {
		return err
	}

	threadID, err := threads.CreateOrGetExistingGitHubPullRequest(ctx, repo.DBID(), repo.DBExternalRepo(), threads.CreateChangesetData{
		BaseRefName: defaultBranch.AbbrevName(),
		HeadRefName: branchName,
		Title:       thread.Title(),
		Body:        thread.Body() + fmt.Sprintf(`\n\n<img src="https://about.sourcegraph.com/sourcegraph-mark.png" width=12 height=12> Campaign: [%s](#)`, campaignName),
	})
	if err != nil {
		return err
	}
	if err := addRemoveThreadsToFromCampaign(ctx, graphqlbackend.MarshalCampaignID(campaignID), []graphql.ID{graphqlbackend.MarshalThreadID(threadID)}, nil); err != nil {
		return err
	}

	diagConnection, err := thread.Diagnostics(ctx, &graphqlutil.ConnectionArgs{})
	if err != nil {
		return err
	}
	diags, err := diagConnection.Nodes(ctx)
	if err != nil {
		return err
	}
	if _, err := graphqlbackend.ThreadDiagnostics.AddDiagnosticsToThread(ctx, &graphqlbackend.AddDiagnosticsToThreadArgs{Thread: graphqlbackend.MarshalThreadID(threadID), RawDiagnostics: toRawDiagnosticsFromGQL(diags)}); err != nil {
		return err
	}

	return nil
}

func (x *rulesExecutor) executeRules(ctx context.Context, campaignID int64, campaignName string) error {
	return x.syncThreads(ctx, campaignID, campaignName)
}

func toRawDiagnosticsFromGQL(diags []graphqlbackend.Diagnostic) []string {
	rawDiagnostics := make([]string, len(diags))
	for i, d := range diags {
		b, err := json.Marshal(d.(diagnostics.GQLDiagnostic))
		if err != nil {
			panic(err)
		}
		rawDiagnostics[i] = string(b)
	}
	return rawDiagnostics
}
