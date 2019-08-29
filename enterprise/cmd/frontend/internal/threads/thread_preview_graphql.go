package threads

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/diagnostics"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func NewGQLThreadPreview(input graphqlbackend.CreateThreadInput, repoComparison graphqlbackend.RepositoryComparison) graphqlbackend.ThreadPreview {
	return &gqlThreadPreview{input: input, repoComparison: repoComparison}
}

type gqlThreadPreview struct {
	input          graphqlbackend.CreateThreadInput
	repoComparison graphqlbackend.RepositoryComparison
}

func (v *gqlThreadPreview) Internal_Input() graphqlbackend.CreateThreadInput { return v.input }

func (v *gqlThreadPreview) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return graphqlbackend.RepositoryByID(ctx, v.input.Repository)
}

func (v *gqlThreadPreview) Internal_RepositoryID() api.RepoID {
	dbID, err := graphqlbackend.UnmarshalRepositoryID(v.input.Repository)
	if err != nil {
		panic(err)
	}
	return api.RepoID(dbID)
}

func (v *gqlThreadPreview) Title() string { return v.input.Title }

func (v *gqlThreadPreview) IsDraft() bool { return v.input.Draft != nil && *v.input.Draft }

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
	var diags []graphqlbackend.Diagnostic
	if v.input.RawDiagnostics != nil {
		diags = make([]graphqlbackend.Diagnostic, len(*v.input.RawDiagnostics))
		for i, rd := range *v.input.RawDiagnostics {
			var d diagnostics.GQLDiagnostic
			if err := json.Unmarshal([]byte(rd), &d); err != nil {
				return nil, err
			}
			diags[i] = d
		}
	}
	return diagnostics.ConstConnection(diags), nil
}

func (v *gqlThreadPreview) Kind(ctx context.Context) (graphqlbackend.ThreadKind, error) {
	if v.repoComparison != nil {
		return graphqlbackend.ThreadKindChangeset, nil
	}
	return graphqlbackend.ThreadKindIssue, nil
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

func (v *gqlThreadPreview) Labels(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.LabelConnection, error) {
	return graphqlbackend.EmptyLabelConnection, nil // empty for now
}

func (v *gqlThreadPreview) Assignees(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.ActorConnection, error) {
	// TODO!(sqs): hack, get code owners
	//

	// repoComparison, err := v.RepositoryComparison(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	repo, err := v.Repository(ctx)
	if err != nil {
		return nil, err
	}
	defaultBranch, err := repo.DefaultBranch(ctx)
	if err != nil {
		return nil, err
	}
	commit, err := defaultBranch.Target().Commit(ctx)
	if err != nil {
		return nil, err
	}
	if commit == nil {
		return graphqlbackend.ActorConnection{}, nil
	}
	person := commit.Author().Person()
	user, err := person.User(ctx)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return graphqlbackend.ActorConnection{graphqlbackend.Actor{User: user}}, nil
	}

	username, err := person.Name(ctx)
	if err != nil {
		return nil, err
	}
	displayName, err := person.DisplayName(ctx)
	if err != nil {
		return nil, err
	}
	email := person.Email()
	return graphqlbackend.ActorConnection{
		graphqlbackend.Actor{
			ExternalActor: &graphqlbackend.ExternalActor{
				Username_:    username,
				DisplayName_: &displayName,
				URL_:         "mailto:" + email,
			},
		},
	}, nil
}

func (v *gqlThreadPreview) InternalID() (string, error) {
	b, err := json.Marshal(v.input)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])[:32], nil
}
