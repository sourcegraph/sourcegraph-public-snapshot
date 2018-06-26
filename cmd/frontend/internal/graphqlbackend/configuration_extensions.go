package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func (r *configurationMutationResolver) UpdateExtension(ctx context.Context, args *struct {
	Extension   *graphql.ID
	ExtensionID *string
	Enabled     *bool
	Remove      *bool
}) (*EmptyResponse, error) {
	var extensionID string
	switch {
	case args.Extension != nil && args.ExtensionID != nil:
		return nil, errors.New("either extension or extensionID must be set, not both")

	case args.Extension == nil && args.ExtensionID == nil:
		return nil, errors.New("exactly 1 of extension or extensionID must be set")

	case args.Enabled != nil && args.Remove != nil && *args.Remove:
		return nil, errors.New("either enabled or remove must be set, not both")

	case args.Enabled == nil && args.Remove == nil:
		return nil, errors.New("exactly 1 of enabled or remove must be set")

	case args.Extension != nil:
		extension, err := registryExtensionByID(ctx, *args.Extension)
		if err != nil {
			return nil, err
		}
		if ok, err := extension.ViewerCanConfigure(ctx); err != nil {
			return nil, err
		} else if !ok {
			// This is NOT for security, it is only to prevent confusion by users who try to enable
			// extensions that won't work for them. They can still manually edit their settings to
			// enable them (and presumably they would not work if viewerCanConfigure == false).
			return nil, errors.New("viewer can't use extension")
		}
		extensionID = extension.ExtensionID()

	case args.ExtensionID != nil:
		extensionID = *args.ExtensionID
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
	return &EmptyResponse{}, nil
}
