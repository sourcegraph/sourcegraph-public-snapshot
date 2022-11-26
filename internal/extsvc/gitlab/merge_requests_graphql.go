package gitlab

import (
	"context"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/graphql"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (c *Client) getMergeRequestGraphQL(ctx context.Context, project *Project, iid ID) (*MergeRequest, error) {
	time.Sleep(c.rateLimitMonitor.RecommendedWaitForBackgroundOp(1))

	resp, err := graphql.GetMergeRequest(ctx, c.gqlClient, project.PathWithNamespace, string(iid))
	if err != nil {
		return nil, errors.Wrap(err, "querying merge request")
	}

	mr, err := mergeRequestFromGraphQL(&resp.Project.MergeRequest.MergeRequest)
	if err != nil {
		return nil, errors.Wrap(err, "converting merge request")
	}

	labels := make([]string, len(resp.Project.MergeRequest.Labels.Nodes))
	for i, label := range resp.Project.MergeRequest.Labels.Nodes {
		labels[i] = label.Title
	}
	mr.Labels = newGraphQLPaginatedResult(ctx, &resp.Project.MergeRequest.Labels.PageInfo.PageInfo, labels, func(ctx context.Context, cursor string) ([]string, *graphql.PageInfo, error) {
		resp, err := graphql.GetMergeRequestLabelsPage(ctx, c.gqlClient, project.PathWithNamespace, string(iid), cursor)
		if err != nil {
			return nil, nil, err
		}

		labels := make([]string, len(resp.Project.MergeRequest.Labels.Nodes))
	})

	return mr, nil
}

func mergeRequestFromGraphQL(raw *graphql.MergeRequest) (*MergeRequest, error) {
	if raw == nil {
		return nil, nil
	}

	id, err := graphql.ExtractRestID(raw.Id)
	if err != nil {
		return nil, errors.Wrap(err, "parsing ID")
	}

	parsedIID, err := strconv.Atoi(raw.Iid)
	if err != nil {
		return nil, errors.Wrap(err, "parsing IID")
	}

	headPipeline, err := pipelineFromGraphQL(&raw.HeadPipeline.Pipeline)
	if err != nil {
		return nil, errors.Wrap(err, "parsing head pipeline")
	}

	user, err := userFromGraphQL(&raw.Author)
	if err != nil {
		return nil, errors.Wrap(err, "parsing author")
	} else if user == nil {
		return nil, errors.New("no author provided in merge request")
	}

	return &MergeRequest{
		ID:              ID(id),
		IID:             ID(parsedIID),
		ProjectID:       ID(raw.ProjectId),
		SourceProjectID: ID(raw.SourceProjectId),
		Title:           raw.Title,
		Description:     raw.Description,
		State:           MergeRequestState(raw.State),
		CreatedAt:       Time{raw.CreatedAt},
		UpdatedAt:       Time{raw.UpdatedAt},
		MergedAt:        NewTime(graphql.OptionalTime(raw.MergedAt)),
		HeadPipeline:    headPipeline,
		SourceBranch:    raw.SourceBranch,
		TargetBranch:    raw.TargetBranch,
		WebURL:          raw.WebUrl,
		WorkInProgress:  raw.Draft,
		Draft:           raw.Draft,
		Author:          *user,
		DiffRefs: DiffRefs{
			BaseSHA:  raw.DiffRefs.BaseSha,
			HeadSHA:  raw.DiffRefs.HeadSha,
			StartSHA: raw.DiffRefs.StartSha,
		},
		SourceProjectNamespace: raw.SourceProject.Namespace.Name,
	}, nil
}

func pipelineFromGraphQL(raw *graphql.Pipeline) (*Pipeline, error) {
	if raw == nil {
		return nil, nil
	}

	id, err := graphql.ExtractRestID(raw.Id)
	if err != nil {
		return nil, errors.Wrap(err, "parsing ID")
	}

	return &Pipeline{
		ID:        ID(id),
		SHA:       raw.Sha,
		Ref:       raw.Ref,
		Status:    PipelineStatus(raw.Status),
		WebURL:    raw.Path,
		CreatedAt: Time{raw.CreatedAt},
		UpdatedAt: Time{raw.UpdatedAt},
	}, nil
}

func userFromGraphQL(raw graphql.User) (*User, error) {
	if raw == nil {
		return nil, nil
	}

	id, err := graphql.ExtractRestID(raw.GetId())
	if err != nil {
		return nil, errors.Wrap(err, "parsing ID")
	}

	return &User{
		ID:        int32(id),
		Name:      raw.GetName(),
		Username:  raw.GetUsername(),
		Email:     raw.GetPublicEmail(),
		State:     string(raw.GetState()),
		AvatarURL: raw.GetAvatarUrl(),
		WebURL:    raw.GetWebUrl(),
		// There's no way to get identities through GraphQL as at the time of
		// writing (version 15.6).
		Identities: []Identity{},
	}, err
}
