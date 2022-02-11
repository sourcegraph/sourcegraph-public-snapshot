package graphql

import (
	"context"
	"encoding/json"
	"math"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *QueryResolver) DocumentationPage(ctx context.Context, args *gql.LSIFDocumentationPageArgs) (gql.DocumentationPageResolver, error) {
	page, err := r.resolver.DocumentationPage(ctx, args.PathID)
	if err != nil {
		return nil, err
	}
	if page == nil {
		return nil, errors.New("page not found")
	}
	tree, err := json.Marshal(page.Tree)
	if err != nil {
		return nil, err
	}
	return &DocumentationPageResolver{tree: gql.JSONValue{Value: string(tree)}}, nil
}

type DocumentationPageResolver struct {
	tree gql.JSONValue
}

func (r *DocumentationPageResolver) Tree() gql.JSONValue {
	return r.tree
}

func (r *QueryResolver) DocumentationPathInfo(ctx context.Context, args *gql.LSIFDocumentationPathInfoArgs) (gql.JSONValue, error) {
	var maxDepth = 1
	if args.MaxDepth != nil {
		maxDepth = int(*args.MaxDepth)
		if maxDepth < 0 {
			maxDepth = int(math.MaxInt32)
		}
	}
	ignoreIndex := false
	if args.IgnoreIndex != nil {
		ignoreIndex = *args.IgnoreIndex
	}

	var get func(pathID string, depth int) (*DocumentationPathInfoResult, error)
	get = func(pathID string, depth int) (*DocumentationPathInfoResult, error) {
		pathInfo, err := r.resolver.DocumentationPathInfo(ctx, pathID)
		if err != nil {
			return nil, err
		}
		if pathInfo == nil {
			return nil, nil
		}
		children := []DocumentationPathInfoResult{}
		if depth < maxDepth {
			if !ignoreIndex || ignoreIndex && !pathInfo.IsIndex {
				depth++
			}
			for _, childPathID := range pathInfo.Children {
				child, err := get(childPathID, depth)
				if err != nil {
					return nil, err
				}
				if child != nil {
					children = append(children, *child)
				}
			}
		}
		return &DocumentationPathInfoResult{
			PathID:   pathInfo.PathID,
			IsIndex:  pathInfo.IsIndex,
			Children: children,
		}, nil
	}

	result, err := get(args.PathID, -1)
	if err != nil {
		return gql.JSONValue{}, err
	}
	if result == nil {
		return gql.JSONValue{}, errors.New("page not found")
	}

	data, err := json.Marshal(result)
	if err != nil {
		return gql.JSONValue{}, err
	}
	return gql.JSONValue{Value: string(data)}, nil
}

// DocumentationPathInfoResult describes a single documentation page path, what is located there
// and what pages are below it.
type DocumentationPathInfoResult struct {
	// The pathID for this page/entry.
	PathID string `json:"pathID"`

	// IsIndex tells if the page at this path is an empty index page whose only purpose is to describe
	// all the pages below it.
	IsIndex bool `json:"isIndex"`

	// Children is a list of the children page paths immediately below this one.
	Children []DocumentationPathInfoResult `json:"children"`
}

func (r *QueryResolver) DocumentationDefinitions(ctx context.Context, args *gql.LSIFQueryDocumentationArgs) (gql.LocationConnectionResolver, error) {
	locations, err := r.resolver.DocumentationDefinitions(ctx, args.PathID)
	if err != nil {
		return nil, err
	}

	return NewLocationConnectionResolver(locations, nil, r.locationResolver), nil
}

func (r *QueryResolver) DocumentationReferences(ctx context.Context, args *gql.LSIFPagedQueryDocumentationArgs) (gql.LocationConnectionResolver, error) {
	limit := derefInt32(args.First, DefaultReferencesPageSize)
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}
	cursor, err := graphqlutil.DecodeCursor(args.After)
	if err != nil {
		return nil, err
	}

	locations, cursor, err := r.resolver.DocumentationReferences(ctx, args.PathID, limit, cursor)
	if err != nil {
		return nil, err
	}

	return NewLocationConnectionResolver(locations, strPtr(cursor), r.locationResolver), nil
}
