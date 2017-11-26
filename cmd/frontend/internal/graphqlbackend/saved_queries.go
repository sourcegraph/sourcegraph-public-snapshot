package graphqlbackend

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/jsonx"
)

type savedQueryResolver struct {
	subject     *configurationSubject
	index       int
	description string
	query       searchQuery
}

func (r savedQueryResolver) Subject() *configurationSubject { return r.subject }

func (r savedQueryResolver) Index() int32 { return int32(r.index) }

func (r savedQueryResolver) Description() string { return r.description }

func (r savedQueryResolver) Query() *searchQuery { return &r.query }

// configSavedQuery is the JSON shape of a saved query entry in the JSON configuration
// (i.e., an entry in the {"search.savedQueries": [...]} array).
type configSavedQuery struct {
	Description string `json:"description"`
	Query       string `json:"query"`
	ScopeQuery  string `json:"scopeQuery"`
}

// partialConfigSavedQueries is the JSON configuration shape, including only the
// search.savedQueries section.
type partialConfigSavedQueries struct {
	SavedQueries []configSavedQuery `json:"search.savedQueries"`
}

func (r *rootResolver) SavedQueries(ctx context.Context) ([]*savedQueryResolver, error) {
	configSubjects, err := r.Configuration().Subjects(ctx)
	if err != nil {
		return nil, err
	}

	var savedQueries []*savedQueryResolver
	for _, subject := range configSubjects {
		var config partialConfigSavedQueries
		if err := subject.readConfiguration(ctx, &config); err != nil {
			return nil, err
		}
		for i, e := range config.SavedQueries {
			savedQueries = append(savedQueries, &savedQueryResolver{
				subject:     subject,
				index:       i,
				description: e.Description,
				query:       searchQuery{query: e.Query, scopeQuery: e.ScopeQuery},
			})
		}
	}
	return savedQueries, nil
}

func (r *configurationMutationResolver) CreateSavedQuery(ctx context.Context, args *struct {
	Description string
	Query       string
	ScopeQuery  string
}) (*savedQueryResolver, error) {
	var index int
	_, err := r.doUpdateConfiguration(ctx, func(oldConfig string) (edits []jsonx.Edit, err error) {
		// Compute the index so we can return it to the caller.
		var config partialConfigSavedQueries
		if err := json.Unmarshal(normalizeJSON(oldConfig), &config); err != nil {
			return nil, err
		}
		index = len(config.SavedQueries)

		value := configSavedQuery{
			Description: args.Description,
			Query:       args.Query,
			ScopeQuery:  args.ScopeQuery,
		}
		edits, _, err = jsonx.ComputePropertyEdit(oldConfig, jsonx.MakePath("search.savedQueries", -1), value, nil, formatOptions)
		return edits, err
	})
	if err != nil {
		return nil, err
	}
	return &savedQueryResolver{
		subject:     r.subject,
		index:       index,
		description: args.Description,
		query:       searchQuery{query: args.Query, scopeQuery: args.ScopeQuery},
	}, nil
}

func (r *configurationMutationResolver) UpdateSavedQuery(ctx context.Context, args *struct {
	Index       int32
	Description *string
	Query       *string
	ScopeQuery  *string
}) (*savedQueryResolver, error) {
	// Do a field-by-field update so we preserve comments and any other unrecognized fields
	// in the object.
	fieldUpdates := map[string]string{}
	if args.Description != nil {
		fieldUpdates["description"] = *args.Description
	}
	if args.Query != nil {
		fieldUpdates["query"] = *args.Query
	}
	if args.ScopeQuery != nil {
		fieldUpdates["scopeQuery"] = *args.ScopeQuery
	}
	for propertyName, value := range fieldUpdates {
		id, err := r.doUpdateConfiguration(ctx, func(oldConfig string) (edits []jsonx.Edit, err error) {
			keyPath := jsonx.MakePath("search.savedQueries", int(args.Index), propertyName)
			edits, _, err = jsonx.ComputePropertyEdit(oldConfig, keyPath, value, nil, formatOptions)
			return edits, err
		})
		if err != nil {
			return nil, err
		}
		r.input.LastID = &id
	}

	// Get final saved query value to return.
	var config partialConfigSavedQueries
	if err := r.subject.readConfiguration(ctx, &config); err != nil {
		return nil, err
	}
	savedQuery := config.SavedQueries[int(args.Index)]
	return &savedQueryResolver{
		subject:     r.subject,
		index:       int(args.Index),
		description: savedQuery.Description,
		query:       searchQuery{query: savedQuery.Query, scopeQuery: savedQuery.ScopeQuery},
	}, nil
}

func (r *configurationMutationResolver) DeleteSavedQuery(ctx context.Context, args *struct {
	Index int32
}) (*EmptyResponse, error) {
	_, err := r.doUpdateConfiguration(ctx, func(oldConfig string) (edits []jsonx.Edit, err error) {
		edits, _, err = jsonx.ComputePropertyRemoval(oldConfig, jsonx.MakePath("search.savedQueries", int(args.Index)), formatOptions)
		return edits, err
	})
	if err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}
