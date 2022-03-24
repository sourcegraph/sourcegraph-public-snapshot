package settings

import (
	"context"
	"reflect"
	"sort"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/schema"
)

type settingsLoader func(ctx context.Context, db database.DB, v *schema.Settings) error

func ForActor(ctx context.Context, db database.DB) (_ *schema.Settings, err error) {
	tr, ctx := trace.New(ctx, "settings.ForActor", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	settingsLoaders := []settingsLoader{defaultSettingsLoader, siteSettingsLoader}

	a := actor.FromContext(ctx)

	// If we have an authenticated user, we add the users orgs' settings and the users settings as well.
	if a.IsAuthenticated() {
		orgs, err := db.Orgs().GetByUserID(ctx, a.UID)
		if err != nil {
			return nil, err
		}
		// Stable-sort the orgs so that the priority of their settings is stable.
		sort.Slice(orgs, func(i, j int) bool {
			return orgs[i].ID < orgs[j].ID
		})
		// Apply the user's orgs' settings.
		for _, org := range orgs {
			settingsLoaders = append(settingsLoaders, orgSettingsLoader(org.ID))
		}
		// Apply the user's own settings last (it has highest priority).
		settingsLoaders = append(settingsLoaders, userSettingsLoader(a.UID))
	}

	return finalTyped(ctx, db, settingsLoaders)
}

func finalTyped(ctx context.Context, db database.DB, settingsLoaders []settingsLoader) (*schema.Settings, error) {
	allSettings := make([]*schema.Settings, len(settingsLoaders))

	// Each LatestSettings is a roundtrip to the database. So we do the requests concurrently.
	bounded := goroutine.NewBounded(8)
	for i := range settingsLoaders {
		i := i
		bounded.Go(func() error {
			var settings schema.Settings
			err := settingsLoaders[i](ctx, db, &settings)
			if err != nil {
				return err
			}

			// if settings == (schema.Settings{}) {
			// 	return nil
			// }

			allSettings[i] = &settings

			return nil
		})
	}

	if err := bounded.Wait(); err != nil {
		return nil, err
	}

	var merged *schema.Settings
	for _, subjectSettings := range allSettings {
		merged = mergeSettingsLeft(merged, subjectSettings)
	}
	return merged, nil
}

var builtinExtensions = map[string]bool{
	"sourcegraph/apex":       true,
	"sourcegraph/clojure":    true,
	"sourcegraph/cobol":      true,
	"sourcegraph/cpp":        true,
	"sourcegraph/csharp":     true,
	"sourcegraph/cuda":       true,
	"sourcegraph/dart":       true,
	"sourcegraph/elixir":     true,
	"sourcegraph/erlang":     true,
	"sourcegraph/git-extras": true,
	"sourcegraph/go":         true,
	"sourcegraph/graphql":    true,
	"sourcegraph/groovy":     true,
	"sourcegraph/haskell":    true,
	"sourcegraph/java":       true,
	"sourcegraph/jsonnet":    true,
	"sourcegraph/kotlin":     true,
	"sourcegraph/lisp":       true,
	"sourcegraph/lua":        true,
	"sourcegraph/ocaml":      true,
	"sourcegraph/pascal":     true,
	"sourcegraph/perl":       true,
	"sourcegraph/php":        true,
	"sourcegraph/powershell": true,
	"sourcegraph/protobuf":   true,
	"sourcegraph/python":     true,
	"sourcegraph/r":          true,
	"sourcegraph/ruby":       true,
	"sourcegraph/rust":       true,
	"sourcegraph/scala":      true,
	"sourcegraph/shell":      true,
	"sourcegraph/swift":      true,
	"sourcegraph/tcl":        true,
	"sourcegraph/thrift":     true,
	"sourcegraph/typescript": true,
	"sourcegraph/verilog":    true,
	"sourcegraph/vhdl":       true,
}

// FilterRemoteExtensions is called to filter the list of extensions retrieved from the remote
// registry before the list is used by any other part of the application.
//
// It can be overridden to use custom logic.
var FilterRemoteExtensions = func(extensions []string) []string {
	// By default, all remote extensions are allowed.
	return extensions
}

func defaultSettingsLoader(_ context.Context, _ database.DB, v *schema.Settings) error {
	extensionIDs := []string{}
	for id := range builtinExtensions {
		extensionIDs = append(extensionIDs, id)
	}
	extensionIDs = FilterRemoteExtensions(extensionIDs)
	extensions := map[string]bool{}
	for _, id := range extensionIDs {
		extensions[id] = true
	}

	(*v) = schema.Settings{
		ExperimentalFeatures: &schema.SettingsExperimentalFeatures{},
		Extensions:           extensions,
	}

	return nil
}

func siteSettingsLoader(ctx context.Context, db database.DB, v *schema.Settings) error {
	settings, err := db.Settings().GetLatest(ctx, api.SettingsSubject{Site: true})
	if err != nil {
		return err
	}
	return jsonc.Unmarshal(settings.Contents, v)
}

func userSettingsLoader(userID int32) func(ctx context.Context, db database.DB, v *schema.Settings) error {
	return func(ctx context.Context, db database.DB, v *schema.Settings) error {
		settings, err := db.Settings().GetLatest(ctx, api.SettingsSubject{User: &userID})
		if err != nil {
			return err
		}
		return jsonc.Unmarshal(settings.Contents, v)
	}
}

func orgSettingsLoader(orgID int32) func(ctx context.Context, db database.DB, v *schema.Settings) error {
	return func(ctx context.Context, db database.DB, v *schema.Settings) error {
		settings, err := db.Settings().GetLatest(ctx, api.SettingsSubject{Org: &orgID})
		if err != nil {
			return err
		}
		return jsonc.Unmarshal(settings.Contents, v)
	}
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
