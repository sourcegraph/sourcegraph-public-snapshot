package graphqlbackend

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/group"
	"github.com/sourcegraph/sourcegraph/schema"
)

// settingsCascade implements the GraphQL type SettingsCascade (and the deprecated type ConfigurationCascade).
//
// It resolves settings from multiple sources.  When there is overlap between values, they will be
// merged in the following cascading order (first is lowest precedence):
//
// - Global site settings
// - Organization settings
// - Current user settings
type settingsCascade struct {
	db database.DB
	// At most 1 of these fields is set.
	unauthenticatedActor bool
	subject              *settingsSubject
}

var mockSettingsCascadeSubjects func() ([]*settingsSubject, error)

func (r *settingsCascade) Subjects(ctx context.Context) ([]*settingsSubject, error) {
	if mockSettingsCascadeSubjects != nil {
		return mockSettingsCascadeSubjects()
	}

	subjects := []*settingsSubject{{defaultSettings: &defaultSettingsResolver{db: r.db, gqlID: singletonDefaultSettingsGQLID}}, {site: &siteResolver{db: r.db, gqlID: singletonSiteGQLID}}}

	if r.unauthenticatedActor {
		return subjects, nil
	}

	switch {
	case r.subject.site != nil:
		// Nothing more to do.

	case r.subject.org != nil:
		subjects = append(subjects, r.subject)

	case r.subject.user != nil:
		orgs, err := r.db.Orgs().GetByUserID(ctx, r.subject.user.user.ID)
		if err != nil {
			return nil, err
		}
		// Stable-sort the orgs so that the priority of their settings is stable.
		sort.Slice(orgs, func(i, j int) bool {
			return orgs[i].ID < orgs[j].ID
		})
		// Apply the user's orgs' settings.
		for _, org := range orgs {
			subjects = append(subjects, &settingsSubject{org: &OrgResolver{db: r.db, org: org}})
		}
		// Apply the user's own settings last (it has highest priority).
		subjects = append(subjects, r.subject)

	default:
		return nil, errUnknownSettingsSubject
	}

	return subjects, nil
}

func (r *settingsCascade) Final(ctx context.Context) (string, error) {
	settings, err := r.finalTyped(ctx)
	if err != nil {
		return "", err
	}

	settingsBytes, err := json.Marshal(settings)
	return string(settingsBytes), err
}

func (r *settingsCascade) finalTyped(ctx context.Context) (*schema.Settings, error) {
	subjects, err := r.Subjects(ctx)
	if err != nil {
		return nil, err
	}

	// Each LatestSettings is a roundtrip to the database. So we do the requests concurrently.
	g := group.NewWithResults[*schema.Settings]().WithContext(ctx).WithMaxConcurrency(8)
	for _, subject := range subjects {
		subject := subject
		g.Go(func(ctx context.Context) (*schema.Settings, error) {
			settings, err := subject.LatestSettings(ctx)
			if err != nil {
				return nil, err
			}

			if settings == nil {
				return nil, nil
			}

			var unmarshalled schema.Settings
			if err := jsonc.Unmarshal(settings.settings.Contents, &unmarshalled); err != nil {
				return nil, err
			}

			return &unmarshalled, nil
		})
	}

	allSettings, err := g.Wait()
	if err != nil {
		return nil, err
	}

	var merged *schema.Settings
	for _, subjectSettings := range allSettings {
		merged = mergeSettingsLeft(merged, subjectSettings)
	}
	return merged, nil
}

// Deprecated: in the GraphQL API
func (r *settingsCascade) Merged(ctx context.Context) (_ *configurationResolver, err error) {
	tr, ctx := trace.New(ctx, "Merged", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	var messages []string
	s, err := r.Final(ctx)
	if err != nil {
		messages = append(messages, err.Error())
	}
	return &configurationResolver{contents: s, messages: messages}, nil
}

var settingsFieldMergeDepths = map[string]int{
	"SearchScopes":           1,
	"SearchSavedQueries":     1,
	"SearchRepositoryGroups": 1,
	"InsightsDashboards":     1,
	"InsightsAllRepos":       1,
	"Quicklinks":             1,
	"Motd":                   1,
	"Extensions":             1,
	"ExperimentalFeatures":   1,
}

func mergeSettingsLeft(left, right *schema.Settings) *schema.Settings {
	return mergeLeft(reflect.ValueOf(left), reflect.ValueOf(right), 1).Interface().(*schema.Settings)
}

// mergeLeft takes two values of the same type and merges them if possible, ignoring
// any struct fields not listed in deeplyMergedSettingsFieldNames. Its depth parameter
// specifies how many layers deeper to merge, and will be overridden if the struct
// field name matches a name in settingsFieldMergeDepths.
func mergeLeft(left, right reflect.Value, depth int) reflect.Value {
	if left.IsZero() {
		return right
	}

	if right.IsZero() {
		return left
	}

	switch left.Kind() {
	case reflect.Struct:
		if depth <= 0 {
			return right
		}
		leftType := left.Type()
		for i := 0; i < left.NumField(); i++ {
			fieldName := leftType.Field(i).Name
			leftField := left.Field(i)
			rightField := right.Field(i)

			fieldDepth := settingsFieldMergeDepths[fieldName]
			leftField.Set(mergeLeft(leftField, rightField, fieldDepth))
		}
		return left
	case reflect.Map:
		if depth <= 0 {
			return right
		}
		iter := right.MapRange()
		for iter.Next() {
			k := iter.Key()
			rightVal := iter.Value()
			leftVal := left.MapIndex(k)
			if (leftVal != reflect.Value{}) {
				left.SetMapIndex(k, mergeLeft(leftVal, rightVal, depth-1))
			} else {
				left.SetMapIndex(k, rightVal)
			}
		}
		return left
	case reflect.Ptr:
		if depth <= 0 {
			return right
		}
		// Don't decrement depth for pointer deref
		left.Elem().Set(mergeLeft(left.Elem(), right.Elem(), depth))
		return left
	case reflect.Slice:
		if depth <= 0 {
			return right
		}
		return reflect.AppendSlice(left, right)
	}

	// Type is not mergeable, so clobber existing value
	return right
}

func (r schemaResolver) ViewerSettings(ctx context.Context) (*settingsCascade, error) {
	user, err := CurrentUser(ctx, r.db)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &settingsCascade{db: r.db, unauthenticatedActor: true}, nil
	}
	return &settingsCascade{db: r.db, subject: &settingsSubject{user: user}}, nil
}

// Deprecated: in the GraphQL API
func (r *schemaResolver) ViewerConfiguration(ctx context.Context) (*settingsCascade, error) {
	return newSchemaResolver(r.db, r.gitserverClient).ViewerSettings(ctx)
}
