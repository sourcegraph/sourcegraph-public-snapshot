package graphqlbackend

import (
	"context"
	"errors"
	"math/rand"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func (r *configurationMutationResolver) UpdateExtension(ctx context.Context, args *struct {
	Extension   *graphql.ID
	ExtensionID *string
	Enabled     *bool
	Remove      *bool
	Edit        *configurationEdit
}) (*updateExtensionConfigurationResult, error) {
	if envvar.InsecureDevMode() {
		// Simulate latency in dev mode so that we feel the pain of our high-latency users.
		n := rand.Intn(200)
		time.Sleep(time.Duration(200+n) * time.Millisecond)
	}

	var extensionID string
	switch {
	case args.Extension != nil && args.ExtensionID != nil:
		return nil, errors.New("either extension or extensionID must be set, not both")

	case args.Extension == nil && args.ExtensionID == nil:
		return nil, errors.New("exactly 1 of extension or extensionID must be set")

	case (args.Enabled != nil && args.Remove != nil && *args.Remove) || (args.Enabled != nil && args.Edit != nil) || (args.Remove != nil && *args.Remove && args.Edit != nil):
		return nil, errors.New("either enabled, remove, or edit must be set, not all")

	case args.Enabled == nil && args.Remove == nil && args.Edit == nil:
		return nil, errors.New("exactly 1 of enabled or remove must be set")

	case args.Extension != nil:
		extension, err := registryExtensionByID(ctx, *args.Extension)
		if err != nil {
			return nil, err
		}
		extensionID = extension.ExtensionID()

	case args.ExtensionID != nil:
		extensionID = *args.ExtensionID
	}

	if args.Edit != nil {
		// Add key path prefix for the extension.
		keyPath := []*keyPathSegment{{Property: strptr("extensions")}, {Property: strptr(extensionID)}}
		keyPath = append(keyPath, args.Edit.KeyPath...)
		args.Edit.KeyPath = keyPath
		if _, err := r.EditConfiguration(ctx, &struct{ Edit *configurationEdit }{Edit: args.Edit}); err != nil {
			return nil, err
		}
		return &updateExtensionConfigurationResult{extensionID: extensionID, subject: r.subject}, nil
	}

	_, err := r.doUpdateConfiguration(ctx, func(oldConfig string) (edits []jsonx.Edit, err error) {
		path := jsonx.MakePath("extensions", extensionID)
		if args.Remove != nil && *args.Remove {
			edits, _, err = jsonx.ComputePropertyRemoval(oldConfig, path, conf.FormatOptions)
			return edits, err
		}
		if args.Enabled != nil {
			var value map[string]interface{} // of schema.ExtensionSettings type, plus extension-specific config
			// Don't clobber any other settings that already exist for the extension.
			var old struct {
				Extensions map[string]map[string]interface{} `json:"extensions"`
			}
			if err := conf.UnmarshalJSON(oldConfig, &old); err != nil {
				return nil, err
			}
			value = old.Extensions[extensionID]
			if value == nil {
				value = map[string]interface{}{}
			}
			if *args.Enabled {
				delete(value, "disabled")
			} else {
				value["disabled"] = true
			}
			edits, _, err = jsonx.ComputePropertyEdit(oldConfig, jsonx.MakePath("extensions", extensionID), value, nil, conf.FormatOptions)
			return edits, err
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return &updateExtensionConfigurationResult{}, nil
}

type updateExtensionConfigurationResult struct {
	extensionID string
	subject     *configurationSubject
}

func (r *updateExtensionConfigurationResult) MergedSettings(ctx context.Context) (*jsonValue, error) {
	return (&configuredExtensionResolver{extensionID: r.extensionID, subject: r.subject}).MergedSettings(ctx)
}
