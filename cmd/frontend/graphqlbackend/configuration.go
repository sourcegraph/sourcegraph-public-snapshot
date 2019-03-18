package graphqlbackend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

type configurationResolver struct {
	contents string
	messages []string // error and warning messages
}

func (r *configurationResolver) Contents() string { return r.contents }

func (r *configurationResolver) Messages() []string {
	if r.messages == nil {
		return []string{}
	}
	return r.messages
}

type configurationMutationGroupInput struct {
	Subject graphql.ID
	LastID  *int32
}

type configurationMutationResolver struct {
	input   *configurationMutationGroupInput
	subject *configurationSubject
}

// ConfigurationMutation defines the Mutation.configuration field.
func (r *schemaResolver) ConfigurationMutation(ctx context.Context, args *struct {
	Input *configurationMutationGroupInput
}) (*configurationMutationResolver, error) {
	subject, err := configurationSubjectByID(ctx, args.Input.Subject)
	if err != nil {
		return nil, err
	}

	// TODO(sqs): support multiple mutations running in a single query that all
	// increment the settings.

	// 🚨 SECURITY: Check whether the viewer can administer this subject (which is equivalent to
	// being able to mutate its configuration).
	if canAdmin, err := subject.ViewerCanAdminister(ctx); err != nil {
		return nil, err
	} else if !canAdmin {
		if !actor.FromContext(ctx).IsAuthenticated() {
			// TODO(sqs): Quick hack to show a less confusing error message when an anon user
			// toggles code coverage on Sourcegraph.com. Make the extension actually show a friendly
			// message and/or implement anon user settings on Sourcegraph.com.
			return nil, errors.New("to toggle coverage or edit settings, you must sign in or sign up")
		}
		return nil, errors.New("viewer is not allowed to edit these settings")
	}

	return &configurationMutationResolver{
		input:   args.Input,
		subject: subject,
	}, nil
}

type updateConfigurationPayload struct{}

func (updateConfigurationPayload) Empty() *EmptyResponse { return nil }

type configurationEdit struct {
	KeyPath                   []*keyPathSegment
	Value                     *jsonValue
	ValueIsJSONCEncodedString bool
}

type keyPathSegment struct {
	Property *string
	Index    *int32
}

func toKeyPath(gqlKeyPath []*keyPathSegment) (jsonx.Path, error) {
	keyPath := make(jsonx.Path, len(gqlKeyPath))
	for i, s := range gqlKeyPath {
		if (s.Property == nil) == (s.Index == nil) {
			return nil, fmt.Errorf("invalid key path segment at index %d: exactly 1 of property and index must be non-null", i)
		}

		var segment jsonx.Segment
		if s.Property != nil {
			segment.IsProperty = true
			segment.Property = *s.Property
		} else {
			segment.Index = int(*s.Index)
		}
		keyPath[i] = segment
	}
	return keyPath, nil
}

func (r *configurationMutationResolver) EditConfiguration(ctx context.Context, args *struct {
	Edit *configurationEdit
}) (*updateConfigurationPayload, error) {
	keyPath, err := toKeyPath(args.Edit.KeyPath)
	if err != nil {
		return nil, err
	}

	remove := args.Edit.Value == nil
	var value interface{}
	if args.Edit.Value != nil {
		value = args.Edit.Value.value
	}
	if args.Edit.ValueIsJSONCEncodedString {
		s, ok := value.(string)
		if !ok {
			return nil, errors.New("value must be a string for valueIsJSONCEncodedString")
		}
		value = json.RawMessage(s)
	}

	return r.editConfiguration(ctx, keyPath, value, remove)
}

func (r *configurationMutationResolver) editConfiguration(ctx context.Context, keyPath jsonx.Path, value interface{}, remove bool) (*updateConfigurationPayload, error) {
	_, err := r.doUpdateConfiguration(ctx, func(oldConfig string) (edits []jsonx.Edit, err error) {
		if remove {
			edits, _, err = jsonx.ComputePropertyRemoval(oldConfig, keyPath, conf.FormatOptions)
		} else {
			edits, _, err = jsonx.ComputePropertyEdit(oldConfig, keyPath, value, nil, conf.FormatOptions)
		}
		return edits, err
	})
	if err != nil {
		return nil, err
	}
	return &updateConfigurationPayload{}, nil
}

func (r *configurationMutationResolver) OverwriteConfiguration(ctx context.Context, args *struct {
	Contents *string
}) (*updateConfigurationPayload, error) {
	_, err := settingsCreateIfUpToDate(ctx, r.subject, r.input.LastID, actor.FromContext(ctx).UID, *args.Contents)
	if err != nil {
		return nil, err
	}
	return &updateConfigurationPayload{}, nil
}

// doUpdateConfiguration is a helper for updating configuration.
func (r *configurationMutationResolver) doUpdateConfiguration(ctx context.Context, computeEdits func(oldConfig string) ([]jsonx.Edit, error)) (idAfterUpdate int32, err error) {
	currentConfig, err := r.getCurrentConfig(ctx)
	if err != nil {
		return 0, err
	}

	edits, err := computeEdits(currentConfig)
	if err != nil {
		return 0, err
	}
	newConfig, err := jsonx.ApplyEdits(currentConfig, edits...)
	if err != nil {
		return 0, err
	}

	// Write mutated settings.
	updatedSettings, err := settingsCreateIfUpToDate(ctx, r.subject, r.input.LastID, actor.FromContext(ctx).UID, newConfig)
	if err != nil {
		return 0, err
	}
	return updatedSettings.ID, nil
}

func (r *configurationMutationResolver) getCurrentConfig(ctx context.Context) (string, error) {
	// Get the settings file whose contents to mutate.
	settings, err := db.Settings.GetLatest(ctx, r.subject.toSubject())
	if err != nil {
		return "", err
	}
	var config string
	if settings != nil && r.input.LastID != nil && settings.ID == *r.input.LastID {
		config = settings.Contents
	} else if settings == nil && r.input.LastID == nil {
		// noop
	} else {
		intOrNull := func(v *int32) string {
			if v == nil {
				return "null"
			}
			return strconv.FormatInt(int64(*v), 10)
		}
		var lastID *int32
		if settings != nil {
			lastID = &settings.ID
		}
		return "", fmt.Errorf("update configuration version mismatch: last ID is %s (mutation wanted %s)", intOrNull(lastID), intOrNull(r.input.LastID))
	}

	return config, nil
}
