package resolvers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver/client"
)

type lsifQueryResolver struct {
	RepoName string
	Commit   graphqlbackend.GitObjectID
	Path     string
}

var _ graphqlbackend.LSIFQueryResolver = &lsifQueryResolver{}

func (r *lsifQueryResolver) Definitions(ctx context.Context, args *graphqlbackend.LSIFQueryPositionArgs) (graphqlbackend.LocationConnectionResolver, error) {
	opt := LocationsQueryOptions{
		Operation: "definitions",
		RepoName:  r.RepoName,
		Commit:    r.Commit,
		Path:      r.Path,
		Line:      args.Line,
		Character: args.Character,
	}

	resolver, err := resolveLocationConnection(ctx, opt)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

func (r *lsifQueryResolver) References(ctx context.Context, args *graphqlbackend.LSIFPagedQueryPositionArgs) (graphqlbackend.LocationConnectionResolver, error) {
	opt := LocationsQueryOptions{
		Operation: "references",
		RepoName:  r.RepoName,
		Commit:    r.Commit,
		Path:      r.Path,
		Line:      args.Line,
		Character: args.Character,
	}
	if args.First != nil {
		opt.Limit = args.First
	}
	if args.After != nil {
		decoded, err := base64.StdEncoding.DecodeString(*args.After)
		if err != nil {
			return nil, err
		}
		nextURL := string(decoded)
		opt.NextURL = &nextURL
	}

	resolver, err := resolveLocationConnection(ctx, opt)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

func (r *lsifQueryResolver) Hover(ctx context.Context, args *graphqlbackend.LSIFQueryPositionArgs) (graphqlbackend.HoverResolver, error) {
	path := fmt.Sprintf("/hover")
	values := url.Values{}
	values.Set("repository", r.RepoName)
	values.Set("commit", string(r.Commit))
	values.Set("path", r.Path)
	values.Set("line", strconv.FormatInt(int64(args.Line), 10))
	values.Set("character", strconv.FormatInt(int64(args.Character), 10))

	payload := struct {
		Text  string    `json:"text"`
		Range lsp.Range `json:"range"`
	}{}

	if err := client.DefaultClient.TraceRequestAndUnmarshalPayload(ctx, "GET", path, values, nil, &payload); err != nil {
		return nil, err
	}

	return &hoverResolver{
		text:     payload.Text,
		lspRange: payload.Range,
	}, nil
}
