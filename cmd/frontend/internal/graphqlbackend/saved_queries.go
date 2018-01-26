package graphqlbackend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	"github.com/sourcegraph/jsonx"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/query-runner/queryrunnerapi"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/randstring"
)

type savedQueryResolver struct {
	key            string
	subject        *configurationSubject
	index          int
	description    string
	query          searchQuery
	showOnHomepage bool
}

func savedQueryByID(ctx context.Context, id graphql.ID) (*savedQueryResolver, error) {
	var spec api.SavedQueryIDSpec
	if err := relay.UnmarshalSpec(id, &spec); err != nil {
		return nil, err
	}

	subjectID, err := configurationSubjectID(spec.Subject)
	if err != nil {
		return nil, err
	}
	subject, err := configurationSubjectByID(ctx, subjectID)
	if err != nil {
		return nil, err
	}

	var config api.PartialConfigSavedQueries
	if err := subject.readConfiguration(ctx, &config); err != nil {
		return nil, err
	}
	for i, e := range config.SavedQueries {
		if e.Key == spec.Key {
			return toSavedQueryResolver(i, subject, e), nil
		}
	}
	return nil, errors.New("saved query not found")
}

func (r savedQueryResolver) ID() graphql.ID {
	var subject api.ConfigurationSubject
	switch {
	case r.subject.user != nil:
		subject.User = &r.subject.user.user.ID
	case r.subject.org != nil:
		subject.Org = &r.subject.org.org.ID
	}
	return marshalSavedQueryID(api.SavedQueryIDSpec{
		Subject: subject,
		Key:     r.key,
	})
}

func marshalSavedQueryID(spec api.SavedQueryIDSpec) graphql.ID {
	return relay.MarshalID("SavedQuery", spec)
}

func unmarshalSavedQueryID(id graphql.ID) (spec api.SavedQueryIDSpec, err error) {
	err = relay.UnmarshalSpec(id, &spec)
	return
}

func (r savedQueryResolver) ShowOnHomepage() bool {
	return r.showOnHomepage
}

func (r savedQueryResolver) Subject() *configurationSubject { return r.subject }

func (r savedQueryResolver) Key() *string {
	if r.key == "" {
		return nil
	}
	return &r.key
}

func (r savedQueryResolver) Index() int32 { return int32(r.index) }

func (r savedQueryResolver) Description() string { return r.description }

func (r savedQueryResolver) Query() *searchQuery { return &r.query }

func toSavedQueryResolver(index int, subject *configurationSubject, entry api.ConfigSavedQuery) *savedQueryResolver {
	return &savedQueryResolver{
		subject:        subject,
		key:            entry.Key,
		index:          index,
		description:    entry.Description,
		query:          searchQuery{query: entry.Query},
		showOnHomepage: entry.ShowOnHomepage,
	}
}

func (r *schemaResolver) SavedQueries(ctx context.Context) ([]*savedQueryResolver, error) {
	configSubjects, err := r.Configuration().Subjects(ctx)
	if err != nil {
		return nil, err
	}

	var savedQueries []*savedQueryResolver
	for _, subject := range configSubjects {
		var config api.PartialConfigSavedQueries
		if err := subject.readConfiguration(ctx, &config); err != nil {
			return nil, err
		}

		// TEMPORARY: perform migration to add unique key to saved queries and remove scope query by adding field to query field.
		if err := r.migrateSavedQueries(ctx, subject, config.SavedQueries); err != nil {
			return nil, err
		}

		for i, e := range config.SavedQueries {
			savedQueries = append(savedQueries, toSavedQueryResolver(i, subject, e))
		}
	}

	return savedQueries, nil
}

func (r *schemaResolver) migrateSavedQueries(ctx context.Context, subject *configurationSubject, savedQueries []api.ConfigSavedQuery) error {
	// Return if all entries have keys.
	needsKey := false
	hasScopeField := false
	for _, e := range savedQueries {
		if e.Key == "" {
			needsKey = true
		}
		if e.ScopeQuery != "" {
			hasScopeField = true
		}
	}
	if !needsKey && !hasScopeField {
		return nil
	}

	settings, err := subject.LatestSettings(ctx)
	if err != nil {
		return err
	}
	if settings == nil {
		return nil // nothing to do
	}

	mutation, err := r.ConfigurationMutation(ctx, &struct {
		Input *configurationMutationGroupInput
	}{
		Input: &configurationMutationGroupInput{LastID: &settings.settings.ID, Subject: subject.ID()},
	})
	if err != nil {
		return err
	}
	_, err = mutation.doUpdateConfiguration(ctx, func(oldConfig string) (allEdits []jsonx.Edit, err error) {
		for i := range savedQueries {
			if savedQueries[i].Key != "" && savedQueries[i].ScopeQuery == "" {
				continue
			}
			savedQueries[i].Query = savedQueries[i].Query + " " + savedQueries[i].ScopeQuery
			savedQueries[i].ScopeQuery = ""
			savedQueries[i].Key = generateUniqueSavedQueryKey(savedQueries)
			edits, _, err := jsonx.ComputePropertyEdit(oldConfig, jsonx.MakePath("search.savedQueries", i), savedQueries[i], nil, formatOptions)
			if err != nil {
				return nil, err
			}
			allEdits = append(allEdits, edits...)
		}
		return allEdits, nil
	})
	return err
}

func (r *configurationMutationResolver) CreateSavedQuery(ctx context.Context, args *struct {
	Description    string
	Query          string
	ShowOnHomepage bool
}) (*savedQueryResolver, error) {
	var index int
	var key string
	_, err := r.doUpdateConfiguration(ctx, func(oldConfig string) (edits []jsonx.Edit, err error) {
		// Compute the index so we can return it to the caller.
		var config api.PartialConfigSavedQueries
		if err := json.Unmarshal(normalizeJSON(oldConfig), &config); err != nil {
			return nil, err
		}
		index = len(config.SavedQueries)
		key = generateUniqueSavedQueryKey(config.SavedQueries)

		value := api.ConfigSavedQuery{
			Key:            key,
			Description:    args.Description,
			Query:          args.Query,
			ShowOnHomepage: args.ShowOnHomepage,
		}
		edits, _, err = jsonx.ComputePropertyEdit(oldConfig, jsonx.MakePath("search.savedQueries", -1), value, nil, formatOptions)
		return edits, err
	})
	if err != nil {
		return nil, err
	}

	// Read new configuration and inform the query-runner.
	var config api.PartialConfigSavedQueries
	if err := r.subject.readConfiguration(ctx, &config); err != nil {
		return nil, err
	}
	go queryrunnerapi.Client.SavedQueryWasCreatedOrUpdated(context.Background(), r.subject.toSubject(), config)

	return &savedQueryResolver{
		subject:        r.subject,
		key:            key,
		index:          index,
		description:    args.Description,
		query:          searchQuery{query: args.Query},
		showOnHomepage: args.ShowOnHomepage,
	}, nil
}

// getSavedQueryIndex returns the index within the config of the saved query with the given key,
// or else an error.
func (r *configurationMutationResolver) getSavedQueryIndex(ctx context.Context, key string) (int, error) {
	var config api.PartialConfigSavedQueries
	if err := r.subject.readConfiguration(ctx, &config); err != nil {
		return 0, err
	}
	for i, e := range config.SavedQueries {
		if e.Key == key {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no saved query in config with key %q", key)
}

func (r *configurationMutationResolver) UpdateSavedQuery(ctx context.Context, args *struct {
	ID             graphql.ID
	Description    *string
	Query          *string
	ShowOnHomepage bool
}) (*savedQueryResolver, error) {
	spec, err := unmarshalSavedQueryID(args.ID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check that args.ID's encoded subject is the same as the configurationMutation
	// resolver's subject, which we've already checked permissions for.
	if err := r.checkArgHasSameSubject(spec.Subject); err != nil {
		return nil, err
	}

	index, err := r.getSavedQueryIndex(ctx, spec.Key)
	if err != nil {
		return nil, err
	}

	// Do a field-by-field update so we preserve comments and any other unrecognized fields
	// in the object.
	fieldUpdates := map[string]interface{}{}
	if args.Description != nil {
		fieldUpdates["description"] = *args.Description
	}

	if args.Query != nil {
		fieldUpdates["query"] = *args.Query
	}

	fieldUpdates["showOnHomepage"] = args.ShowOnHomepage

	for propertyName, value := range fieldUpdates {
		id, err := r.doUpdateConfiguration(ctx, func(oldConfig string) (edits []jsonx.Edit, err error) {
			keyPath := jsonx.MakePath("search.savedQueries", index, propertyName)
			edits, _, err = jsonx.ComputePropertyEdit(oldConfig, keyPath, value, nil, formatOptions)
			return edits, err
		})
		if err != nil {
			return nil, err
		}
		r.input.LastID = &id
	}

	// Get final saved query value to return.
	var config api.PartialConfigSavedQueries
	if err := r.subject.readConfiguration(ctx, &config); err != nil {
		return nil, err
	}
	go queryrunnerapi.Client.SavedQueryWasCreatedOrUpdated(context.Background(), spec.Subject, config)
	return toSavedQueryResolver(index, r.subject, config.SavedQueries[index]), nil
}

func (r *configurationMutationResolver) DeleteSavedQuery(ctx context.Context, args *struct {
	ID graphql.ID
}) (*EmptyResponse, error) {
	spec, err := unmarshalSavedQueryID(args.ID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check that args.ID's encoded subject is the same as the configurationMutation
	// resolver's subject, which we've already checked permissions for.
	if err := r.checkArgHasSameSubject(spec.Subject); err != nil {
		return nil, err
	}

	index, err := r.getSavedQueryIndex(ctx, spec.Key)
	if err != nil {
		return nil, err
	}

	_, err = r.doUpdateConfiguration(ctx, func(oldConfig string) (edits []jsonx.Edit, err error) {
		edits, _, err = jsonx.ComputePropertyRemoval(oldConfig, jsonx.MakePath("search.savedQueries", index), formatOptions)
		return edits, err
	})
	if err != nil {
		return nil, err
	}
	go queryrunnerapi.Client.SavedQueryWasDeleted(context.Background(), spec)
	return &EmptyResponse{}, nil
}

func generateUniqueSavedQueryKey(existing []api.ConfigSavedQuery) string {
	// Avoid collisions.
	used := make(map[string]struct{}, len(existing))
	for _, e := range existing {
		used[e.Key] = struct{}{}
	}

	// Omit I, L, O, C, and U to reduce the likelihood of confusion with digits
	// and accidental obscenity. (Similar to https://en.wikipedia.org/wiki/Base32#Crockford's_Base32.)
	const niceEnglishChars = "ABDEFGHJKMNPQRSTVWXYZabdefghjkmnpqrstvwxyz0123456789"
	const maxIter = 100
	for i := 0; i < maxIter; i++ {
		candidate := randstring.NewLenChars(6, []byte(niceEnglishChars))
		if _, used := used[candidate]; used {
			continue // collision
		}
		return candidate
	}
	panic(fmt.Sprintf("unable to generate unique saved query key after %d iterations (used %d unique keys)", maxIter, len(used)))
}
