package threads

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
)

func NewGQLThreadPreview(input graphqlbackend.CreateThreadInput, repoComparison graphqlbackend.RepositoryComparison) graphqlbackend.ThreadPreview {
	return &gqlThreadPreview{input: input, repoComparison: repoComparison}
}

type gqlThreadPreview struct {
	input          graphqlbackend.CreateThreadInput
	repoComparison graphqlbackend.RepositoryComparison
}

func (v *gqlThreadPreview) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return graphqlbackend.RepositoryByID(ctx, v.input.Repository)
}

func (v *gqlThreadPreview) Title() string { return v.input.Title }

func (v *gqlThreadPreview) Author(ctx context.Context) (*graphqlbackend.Actor, error) {
	user, err := graphqlbackend.CurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	return &graphqlbackend.Actor{User: user}, nil
}

func (v *gqlThreadPreview) Body() string {
	if v.input.Body == nil {
		return ""
	}
	return *v.input.Body
}

func (v *gqlThreadPreview) BodyText() string { return comments.ToBodyText(v.Body()) }

func (v *gqlThreadPreview) BodyHTML() string { return comments.ToBodyHTML(v.Body()) }

func (v *gqlThreadPreview) Diagnostics(context.Context, *graphqlutil.ConnectionArgs) (graphqlbackend.DiagnosticConnection, error) {
	panic("TODO!(sqs)")
}

func (v *gqlThreadPreview) Kind(ctx context.Context) (graphqlbackend.ThreadKind, error) {
	// TODO!(sqs) un-hardcode
	return graphqlbackend.ThreadKindChangeset, nil
}

func (v *gqlThreadPreview) RepositoryComparison(ctx context.Context) (graphqlbackend.RepositoryComparison, error) {
	if v.repoComparison != nil {
		return v.repoComparison, nil
	}

	if v.input.BaseRef == nil && v.input.HeadRef == nil {
		return nil, nil
	}
	repo, err := v.Repository(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlbackend.NewRepositoryComparison(ctx, repo, &graphqlbackend.RepositoryComparisonInput{
		Base: v.input.BaseRef,
		Head: v.input.HeadRef,
	})
}
