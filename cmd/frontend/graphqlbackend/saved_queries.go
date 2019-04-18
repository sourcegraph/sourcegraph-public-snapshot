package graphqlbackend

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/query-runner/queryrunnerapi"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/pkg/randstring"
)

type savedQueryResolver struct {
	key                                 string
	index                               int
	description                         string
	query                               string
	showOnHomepage, notify, notifySlack bool
	ownerKind                           string
	userID, orgID                       *int32
}

func savedQueryByID(ctx context.Context, id graphql.ID) (*savedQueryResolver, error) {
	var spec api.SavedQueryIDSpec
	if err := relay.UnmarshalSpec(id, &spec); err != nil {
		return nil, err
	}

	subjectID, err := settingsSubjectID(spec.Subject)
	if err != nil {
		return nil, err
	}
	subject, err := settingsSubjectByID(ctx, subjectID)
	if err != nil {
		return nil, err
	}

	var config api.PartialConfigSavedQueries
	if err := subject.readSettings(ctx, &config); err != nil {
		return nil, err
	}
	for i, e := range config.SavedQueries {
		if e.Key == spec.Key {
			return toSavedQueryResolver(i, e), nil
		}
	}
	return nil, errors.New("saved query not found")
}

func (r savedQueryResolver) ID() graphql.ID {
	return marshalSavedQueryID(api.SavedQueryIDSpec{Key: r.key})
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

func (r savedQueryResolver) Notify() bool {
	return r.notify
}

func (r savedQueryResolver) NotifySlack() bool {
	return r.notifySlack
}

// DEPRECATED
func (r savedQueryResolver) Subject() *settingsSubject { return nil }

func (r savedQueryResolver) Key() string {
	return r.key
}

// DEPRECATED
func (r savedQueryResolver) Index() int32 { return int32(r.index) }

func (r savedQueryResolver) Description() string { return r.description }

func (r savedQueryResolver) Query() string { return r.query }

func (r savedQueryResolver) OwnerKind() string { return r.ownerKind }
func (r savedQueryResolver) UserID() *int32    { return r.userID }
func (r savedQueryResolver) OrgID() *int32     { return r.orgID }

func toSavedQueryResolver(index int, entry api.ConfigSavedQuery) *savedQueryResolver {
	return &savedQueryResolver{
		key:            entry.Key,
		index:          index,
		description:    entry.Description,
		query:          entry.Query,
		showOnHomepage: entry.ShowOnHomepage,
		notify:         entry.Notify,
		notifySlack:    entry.NotifySlack,
		ownerKind:      entry.UserOrOrg,
		userID:         entry.UserID,
		orgID:          entry.OrgID,
	}
}

func (r *schemaResolver) SavedQueries(ctx context.Context) ([]*savedQueryResolver, error) {
	var savedQueries []*savedQueryResolver
	savedSearches, err := db.SavedSearches.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	for i, savedSearch := range savedSearches {
		savedQueries = append(savedQueries, toSavedQueryResolver(i, savedSearch.Config))
	}

	return savedQueries, nil
}

func (r *settingsMutation) CreateSavedQuery(ctx context.Context, args *struct {
	Description                         string
	Query                               string
	ShowOnHomepage, Notify, NotifySlack bool
	DisableSubscriptionNotifications    bool
}) (*savedQueryResolver, error) {
	var index int
	var key string
	_, err := r.doUpdateSettings(ctx, func(oldConfig string) (edits []jsonx.Edit, err error) {
		// Compute the index so we can return it to the caller.
		var config api.PartialConfigSavedQueries
		if err := jsonc.Unmarshal(oldConfig, &config); err != nil {
			return nil, err
		}
		index = len(config.SavedQueries)
		key = generateUniqueSavedQueryKey(config.SavedQueries)

		value := api.ConfigSavedQuery{
			Key:            key,
			Description:    args.Description,
			Query:          args.Query,
			ShowOnHomepage: args.ShowOnHomepage,
			Notify:         args.Notify,
			NotifySlack:    args.NotifySlack,
		}
		edits, _, err = jsonx.ComputePropertyEdit(oldConfig, jsonx.MakePath("search.savedQueries", -1), value, nil, conf.FormatOptions)
		return edits, err
	})
	if err != nil {
		return nil, err
	}

	// // Read new configuration and inform the query-runner.
	// var config api.PartialConfigSavedQueries
	// if err := r.subject.readSettings(ctx, &config); err != nil {
	// 	return nil, err
	// }
	// go queryrunnerapi.Client.SavedQueryWasCreatedOrUpdated(context.Background(), r.subject.toSubject(), config, args.DisableSubscriptionNotifications)

	return &savedQueryResolver{
		key:            key,
		index:          index,
		description:    args.Description,
		query:          args.Query,
		showOnHomepage: args.ShowOnHomepage,
		notify:         args.Notify,
		notifySlack:    args.NotifySlack,
	}, nil
}

// getSavedQueryIndex returns the index within the config of the saved query with the given key,
// or else an error.
func (r *settingsMutation) getSavedQueryIndex(ctx context.Context, key string) (int, error) {
	var config api.PartialConfigSavedQueries
	if err := r.subject.readSettings(ctx, &config); err != nil {
		return 0, err
	}
	for i, e := range config.SavedQueries {
		if e.Key == key {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no saved query in config with key %q", key)
}

func (r *settingsMutation) UpdateSavedQuery(ctx context.Context, args *struct {
	ID                                  graphql.ID
	Description                         *string
	Query                               *string
	ShowOnHomepage, Notify, NotifySlack bool
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
	fieldUpdates["notify"] = args.Notify
	fieldUpdates["notifySlack"] = args.NotifySlack

	for propertyName, value := range fieldUpdates {
		id, err := r.doUpdateSettings(ctx, func(oldConfig string) (edits []jsonx.Edit, err error) {
			keyPath := jsonx.MakePath("search.savedQueries", index, propertyName)
			edits, _, err = jsonx.ComputePropertyEdit(oldConfig, keyPath, value, nil, conf.FormatOptions)
			return edits, err
		})
		if err != nil {
			return nil, err
		}
		r.input.LastID = &id
	}

	// Get final saved query value to return.
	var config api.PartialConfigSavedQueries
	if err := r.subject.readSettings(ctx, &config); err != nil {
		return nil, err
	}
	// go queryrunnerapi.Client.SavedQueryWasCreatedOrUpdated(context.Background(), spec.Subject, config, false)
	return toSavedQueryResolver(index, config.SavedQueries[index]), nil
}

func (r *settingsMutation) DeleteSavedQuery(ctx context.Context, args *struct {
	ID                               graphql.ID
	DisableSubscriptionNotifications bool
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

	_, err = r.doUpdateSettings(ctx, func(oldConfig string) (edits []jsonx.Edit, err error) {
		edits, _, err = jsonx.ComputePropertyRemoval(oldConfig, jsonx.MakePath("search.savedQueries", index), conf.FormatOptions)
		return edits, err
	})
	if err != nil {
		return nil, err
	}
	// go queryrunnerapi.Client.SavedQueryWasDeleted(context.Background(), spec, args.DisableSubscriptionNotifications)
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

func (r *schemaResolver) SendSavedSearchTestNotification(ctx context.Context, args *struct {
	ID graphql.ID
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Look it up to ensure the actor has access to it.
	if _, err := savedQueryByID(ctx, args.ID); err != nil {
		return nil, err
	}

	spec, err := unmarshalSavedQueryID(args.ID)
	if err != nil {
		return nil, err
	}

	go queryrunnerapi.Client.TestNotification(context.Background(), spec)
	return &EmptyResponse{}, nil
}
