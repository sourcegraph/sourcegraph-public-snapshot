package campaigns

import (
	"context"
	"log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func (GraphQLResolver) CampaignUpdatePreview(ctx context.Context, arg *graphqlbackend.CampaignUpdatePreviewArgs) (graphqlbackend.CampaignUpdatePreview, error) {
	old, err := campaignByID(ctx, arg.Input.Campaign)
	if err != nil {
		return nil, err
	}
	return &gqlCampaignUpdatePreview{old: old, input: arg.Input}, nil
}

type gqlCampaignUpdatePreview struct {
	old   *gqlCampaign
	input graphqlbackend.CampaignUpdatePreviewInput
}

func (v *gqlCampaignUpdatePreview) nameChanged() bool {
	return v.input.Update.Name != nil && v.old.Name() != *v.input.Update.Name
}

func (v *gqlCampaignUpdatePreview) OldName() *string {
	if v.nameChanged() {
		return strPtr(v.old.Name())
	}
	return nil
}

func (v *gqlCampaignUpdatePreview) NewName() *string {
	if v.nameChanged() {
		return v.input.Update.Name
	}
	return nil
}

func (v *gqlCampaignUpdatePreview) startDateChanged() bool {
	return !v.input.Update.StartDate.Equal(v.old.StartDate())
}

func (v *gqlCampaignUpdatePreview) OldStartDate() *graphqlbackend.DateTime {
	if v.startDateChanged() {
		return v.old.StartDate()
	}
	return nil
}

func (v *gqlCampaignUpdatePreview) NewStartDate() *graphqlbackend.DateTime {
	if v.startDateChanged() {
		return v.input.Update.StartDate
	}
	return nil
}

func (v *gqlCampaignUpdatePreview) dueDateChanged() bool {
	return !v.input.Update.DueDate.Equal(v.old.DueDate())
}

func (v *gqlCampaignUpdatePreview) OldDueDate() *graphqlbackend.DateTime {
	if v.dueDateChanged() {
		return v.old.DueDate()
	}
	return nil
}

func (v *gqlCampaignUpdatePreview) NewDueDate() *graphqlbackend.DateTime {
	if v.dueDateChanged() {
		return v.input.Update.DueDate
	}
	return nil
}

func (v *gqlCampaignUpdatePreview) getThreads(ctx context.Context) ([]graphqlbackend.ToThreadOrThreadPreview, error) {
	// TODO!(sqs): dont ignore args
	info := ruleExecutorCampaignInfo{isDraft: v.old.IsDraft()}
	if v.input.Update.Name != nil {
		info.name = *v.input.Update.Name
	} else {
		info.name = v.old.Name()
	}
	if v.input.Update.Body != nil {
		info.body = *v.input.Update.Body
	} else {
		body, err := v.old.Body(ctx)
		if err != nil {
			return nil, err
		}
		info.body = body
	}

	allThreads, err := (&rulesExecutor{
		campaign:      info,
		extensionData: v.input.Update.ExtensionData,
	}).planThreads(ctx)
	if err != nil {
		return nil, err
	}
	return threads.ToThreadOrThreadPreviews(nil, allThreads), nil
}

func (v *gqlCampaignUpdatePreview) Threads(ctx context.Context) (*[]graphqlbackend.ThreadUpdatePreview, error) {
	oldConnection, err := v.old.Threads(ctx, &graphqlbackend.ThreadConnectionArgs{})
	if err != nil {
		return nil, err
	}
	old, err := oldConnection.Nodes(ctx)
	if err != nil {
		return nil, err
	}

	new, err := v.getThreads(ctx)
	if err != nil {
		return nil, err
	}

	// TODO!(sqs): support multiple threads per repo
	mapByRepo := func(threads []graphqlbackend.ToThreadOrThreadPreview) map[api.RepoID]graphqlbackend.ToThreadOrThreadPreview {
		m := make(map[api.RepoID]graphqlbackend.ToThreadOrThreadPreview, len(threads))
		for _, thread := range threads {
			m[thread.Common().Internal_RepositoryID()] = thread
		}
		return m
	}
	oldByRepo := mapByRepo(old)
	newByRepo := mapByRepo(new)

	var results []graphqlbackend.ThreadUpdatePreview
	for repo, old := range oldByRepo {
		if _, ok := newByRepo[repo]; !ok {
			results = append(results, threads.NewGQLThreadUpdatePreviewForDeletion(old.Thread))
		}
	}
	for repo, new := range newByRepo {
		repoComparison, err := new.RepositoryComparison(ctx)
		if err != nil {
			return nil, err
		}
		if old, ok := oldByRepo[repo]; ok {
			if update, err := threads.NewGQLThreadUpdatePreviewForUpdate(ctx, old.Thread, new.ThreadPreview.Internal_Input(), repoComparison); err != nil {
				return nil, err
			} else if update != nil {
				results = append(results, update)
			}
		} else {
			results = append(results, threads.NewGQLThreadUpdatePreviewForCreation(new.ThreadPreview.Internal_Input(), repoComparison))
		}
	}
	return &results, nil
}

func (v *gqlCampaignUpdatePreview) RepositoryComparisons(ctx context.Context) (*[]*graphqlbackend.RepositoryComparisonUpdatePreview, error) {
	old, err := v.old.RepositoryComparisons(ctx)
	if err != nil {
		return nil, err
	}

	newThreads, err := v.getThreads(ctx)
	if err != nil {
		return nil, err
	}
	new := make([]graphqlbackend.RepositoryComparison, len(newThreads))
	for i, t := range newThreads {
		var err error
		new[i], err = t.RepositoryComparison(ctx)
		if err != nil {
			return nil, err
		}
	}

	// TODO!(sqs): doesnt support >1 thread per repo
	mapByRepo := func(cs []graphqlbackend.RepositoryComparison) map[api.RepoID]graphqlbackend.RepositoryComparison {
		byRepo := make(map[api.RepoID]graphqlbackend.RepositoryComparison, len(cs))
		for _, c := range cs {
			byRepo[c.BaseRepository().DBID()] = c
		}
		return byRepo
	}
	oldByRepo := mapByRepo(old)
	newByRepo := mapByRepo(new)

	log.Println("OLD by repo", oldByRepo)
	log.Println("new by repo", newByRepo)

	var results []*graphqlbackend.RepositoryComparisonUpdatePreview
	// TODO!(sqs): do full delta update, not just new/added
	for repo, new := range newByRepo {
		if old, ok := oldByRepo[repo]; !ok {
			// Added.
			results = append(results, &graphqlbackend.RepositoryComparisonUpdatePreview{
				Repository_: new.HeadRepository(),
				New_:        new,
			})
		} else {
			// Check for change in diff.
			oldDiff, err := old.FileDiffs(&graphqlutil.ConnectionArgs{}).RawDiff(ctx)
			if err != nil {
				return nil, err
			}
			newDiff, err := new.FileDiffs(&graphqlutil.ConnectionArgs{}).RawDiff(ctx)
			if err != nil {
				return nil, err
			}

			oldDiff, err = threads.StripDiffPathPrefixes(oldDiff)
			if err != nil {
				return nil, err
			}
			newDiff, err = threads.StripDiffPathPrefixes(newDiff)
			if err != nil {
				return nil, err
			}

			if oldDiff != newDiff {
				results = append(results, &graphqlbackend.RepositoryComparisonUpdatePreview{
					Repository_: new.HeadRepository(),
					Old_:        old,
					New_:        new,
				})
			}
		}
	}
	if len(results) == 0 {
		return nil, nil
	}
	return &results, nil
}
