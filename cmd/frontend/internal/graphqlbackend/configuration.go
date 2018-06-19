package graphqlbackend

import (
	"context"
	"fmt"
	"strconv"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
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

	return &configurationMutationResolver{
		input:   args.Input,
		subject: subject,
	}, nil
}

type updateConfigurationInput struct {
	Property string
	Value    *jsonValue
}

type updateConfigurationPayload struct{}

func (updateConfigurationPayload) Empty() *EmptyResponse { return nil }

func (r *configurationMutationResolver) UpdateConfiguration(ctx context.Context, args *struct {
	Input *updateConfigurationInput
}) (*updateConfigurationPayload, error) {
	config, err := r.getCurrentConfig(ctx)
	if err != nil {
		return nil, err
	}

	remove := args.Input.Value == nil
	var value interface{}
	if args.Input.Value != nil {
		value = args.Input.Value.value
	}

	keyPath := jsonx.PropertyPath(args.Input.Property)
	_, err = r.doUpdateConfiguration(ctx, func(oldConfig string) (edits []jsonx.Edit, err error) {
		if remove {
			edits, _, err = jsonx.ComputePropertyRemoval(config, keyPath, conf.FormatOptions)
		} else {
			edits, _, err = jsonx.ComputePropertyEdit(config, keyPath, value, nil, conf.FormatOptions)
		}
		return edits, err
	})
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
	updatedSettings, err := db.Settings.CreateIfUpToDate(ctx, r.subject.toSubject(), r.input.LastID, actor.FromContext(ctx).UID, newConfig)
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
