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
	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

type lsifQueryResolver struct {
	repoName string
	commit   graphqlbackend.GitObjectID
	path     string
	dump     *lsif.LSIFDump
}

var _ graphqlbackend.LSIFQueryResolver = &lsifQueryResolver{}

func (r *lsifQueryResolver) Definitions(ctx context.Context, args *graphqlbackend.LSIFQueryPositionArgs) (graphqlbackend.LocationConnectionResolver, error) {
	opt := LocationsQueryOptions{
		Operation: "definitions",
		RepoName:  r.repoName,
		Commit:    r.commit,
		Path:      r.path,
		Line:      args.Line,
		Character: args.Character,
		DumpID:    r.dump.ID,
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
		RepoName:  r.repoName,
		Commit:    r.commit,
		Path:      r.path,
		Line:      args.Line,
		Character: args.Character,
		DumpID:    r.dump.ID,
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
	values.Set("repository", r.repoName)
	values.Set("commit", string(r.commit))
	values.Set("path", r.path)
	values.Set("line", strconv.FormatInt(int64(args.Line), 10))
	values.Set("character", strconv.FormatInt(int64(args.Character), 10))
	values.Set("dumpId", strconv.FormatInt(r.dump.ID, 10))

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
