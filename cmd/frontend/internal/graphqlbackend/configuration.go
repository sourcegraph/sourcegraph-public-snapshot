package graphqlbackend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	"github.com/sourcegraph/jsonx"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

type configurationSubject struct {
	org  *orgResolver
	user *userResolver
}

// configurationSubjectByID fetches the configuration subject with the given ID. If the ID
// refers to a node that is not a valid configuration subject, an error is returned.
func configurationSubjectByID(ctx context.Context, id graphql.ID) (*configurationSubject, error) {
	resolver, err := nodeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	actor := actor.FromContext(ctx)
	switch s := resolver.(type) {
	case *userResolver:
		// 🚨 SECURITY: A user may only view or modify their own configuration.
		if actor.UID == "" || s.Auth0ID() == "" || actor.UID != s.Auth0ID() {
			return nil, errors.New("a user may only view or modify their own configuration")
		}
		return &configurationSubject{user: s}, nil

	case *orgResolver:
		// 🚨 SECURITY: Check that the current user is a member of the org.
		if _, err := store.OrgMembers.GetByOrgIDAndUserID(ctx, s.org.ID, actor.UID); err != nil {
			return nil, err
		}
		return &configurationSubject{org: s}, nil

	default:
		return nil, errors.New("bad configuration subject type")
	}

}

func idToConfigurationSubject(id graphql.ID) (sourcegraph.ConfigurationSubject, error) {
	switch relay.UnmarshalKind(id) {
	case "User":
		userID, err := unmarshalUserID(id)
		return sourcegraph.ConfigurationSubject{User: &userID}, err
	case "Org":
		orgID, err := unmarshalOrgID(id)
		return sourcegraph.ConfigurationSubject{Org: &orgID}, err
	default:
		return sourcegraph.ConfigurationSubject{}, errors.New("bad configuration subject type")
	}
}

func configurationSubjectID(subject sourcegraph.ConfigurationSubject) (graphql.ID, error) {
	switch {
	case subject.User != nil:
		return marshalUserID(*subject.User), nil
	case subject.Org != nil:
		return marshalOrgID(*subject.Org), nil
	default:
		return "", errors.New("bad configuration subject type")
	}
}

func configurationSubjectsEqual(a, b sourcegraph.ConfigurationSubject) bool {
	switch {
	case a.User != nil && b.User != nil:
		return *a.User == *b.User
	case a.Org != nil && b.Org != nil:
		return *a.Org == *b.Org
	}
	return false
}

// checkArgHasSameSubject ensures that the subject encoded in args.ID (or similar resolver
// field) is the same as that passed to the configurationMutationResolver. If they are different,
// it returns an error.
//
// 🚨 SECURITY: It is used when a mutation field inside the configurationMutation also accepts an
// ID field that encodes the configuration subject. In that case, it's important to check that the
// subjects are equal to prevent a user from bypassing the permission check to write to the
// configuration of the second ID's subject.
func (r *configurationMutationResolver) checkArgHasSameSubject(argSubject sourcegraph.ConfigurationSubject) error {
	if !configurationSubjectsEqual(r.subject.toSubject(), argSubject) {
		return fmt.Errorf("configuration subject mismatch: %s != %s", r.subject.toSubject(), argSubject)
	}
	return nil
}

func (s *configurationSubject) ToOrg() (*orgResolver, bool) { return s.org, s.org != nil }

func (s *configurationSubject) ToUser() (*userResolver, bool) { return s.user, s.user != nil }

func (s *configurationSubject) toSubject() sourcegraph.ConfigurationSubject {
	switch {
	case s.org != nil:
		return sourcegraph.ConfigurationSubject{Org: &s.org.org.ID}
	case s.user != nil:
		return sourcegraph.ConfigurationSubject{User: &s.user.user.ID}
	default:
		panic("no configuration subject")
	}
}

func (s *configurationSubject) ID() graphql.ID {
	switch {
	case s.org != nil:
		return s.org.ID()
	case s.user != nil:
		return s.user.ID()
	}
	panic("no configuration subject")
}

func (s *configurationSubject) LatestSettings(ctx context.Context) (*settingsResolver, error) {
	switch {
	case s.org != nil:
		return s.org.LatestSettings(ctx)
	case s.user != nil:
		return s.user.LatestSettings(ctx)
	}
	panic("no configuration subject")
}

// readConfiguration unmarshals s's latest settings into v.
func (s *configurationSubject) readConfiguration(ctx context.Context, v interface{}) error {
	settings, err := s.LatestSettings(ctx)
	if err != nil {
		return err
	}
	if settings == nil {
		return nil
	}
	return json.Unmarshal(normalizeJSON(settings.Contents()), &v)
}

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

type configurationCascadeResolver struct{}

func (r *configurationCascadeResolver) Defaults() *configurationResolver {
	return &configurationResolver{
		contents: `// This is the default configuration. Override it with org or user settings.
{
  /* default configuration is empty */
}`,
	}
}

var mockConfigurationCascadeSubjects func() ([]*configurationSubject, error)

func (r *configurationCascadeResolver) Subjects(ctx context.Context) ([]*configurationSubject, error) {
	if mockConfigurationCascadeSubjects != nil {
		return mockConfigurationCascadeSubjects()
	}

	var subjects []*configurationSubject
	if actor := actor.FromContext(ctx); actor.IsAuthenticated() {
		user, err := currentUser(ctx)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, nil // actor might be invalid or refer to since-deleted user
		}

		orgs, err := user.Orgs(ctx)
		if err != nil {
			return nil, err
		}
		// Stable-sort the orgs so that the priority of their configs is stable.
		sort.Slice(orgs, func(i, j int) bool {
			return orgs[i].org.ID < orgs[j].org.ID
		})
		// Apply the user's orgs' configuration.
		for _, org := range orgs {
			subjects = append(subjects, &configurationSubject{org: org})
		}

		// Apply the user's own configuration last (it has highest priority).
		subjects = append(subjects, &configurationSubject{user: user})
	}

	return subjects, nil
}

func (r *configurationCascadeResolver) Merged(ctx context.Context) (*configurationResolver, error) {
	configs := []string{r.Defaults().Contents()}
	subjects, err := r.Subjects(ctx)
	if err != nil {
		return nil, err
	}
	for _, s := range subjects {
		settings, err := s.LatestSettings(ctx)
		if err != nil {
			return nil, err
		}
		if settings != nil {
			configs = append(configs, settings.settings.Contents)
		}
	}

	var messages []string
	merged, err := mergeConfigs(configs)
	if err != nil {
		messages = append(messages, err.Error())
	}
	return &configurationResolver{contents: string(merged), messages: messages}, nil
}

// deeplyMergedConfigFields contains the names of top-level configuration fields whose values should
// be merged if they appear in multiple cascading configurations.
//
// For example, suppose org config is {"a":[1]} and user config is {"a":[2]}. If "a" is NOT a deeply
// merged field, the merged config would be {"a":[2]}. If "a" IS a deeply merged field, then the
// merged config would be {"a":[1,2].}
var deeplyMergedConfigFields = map[string]struct{}{
	"search.scopes":       struct{}{},
	"search.savedQueries": struct{}{},
}

// normalizeJSON converts JSON with comments, trailing commas, and some types of syntax errors into
// standard JSON.
func normalizeJSON(input string) []byte {
	output, _ := jsonx.Parse(string(input), jsonx.ParseOptions{Comments: true, TrailingCommas: true})
	if len(output) == 0 {
		return []byte("{}")
	}
	return output
}

// mergeConfigs merges the specified JSON configs together to produce a single JSON config. The merge
// algorithm is currently rudimentary but eventually it will be similar to VS Code's. The only "smart"
// merging behavior is that described in the documentation for deeplyMergedConfigFields.
func mergeConfigs(jsonConfigStrings []string) ([]byte, error) {
	var errs []error
	merged := map[string]interface{}{}
	for _, s := range jsonConfigStrings {
		var config map[string]interface{}
		if err := json.Unmarshal(normalizeJSON(s), &config); err != nil {
			errs = append(errs, err)
			continue
		}
		for name, value := range config {
			// See if we should deeply merge this field.
			if _, ok := deeplyMergedConfigFields[name]; ok {
				if mv, ok := merged[name].([]interface{}); merged[name] == nil || ok {
					if cv, ok := value.([]interface{}); merged[name] != nil || (value != nil && ok) {
						merged[name] = append(mv, cv...)
						continue
					}
				}
			}

			// Otherwise clobber any existing value.
			merged[name] = value
		}
	}
	out, err := json.Marshal(merged)
	if err != nil {
		errs = append(errs, err)
	}
	if len(errs) == 0 {
		return out, nil
	}
	return out, fmt.Errorf("errors merging configurations: %q", errs)
}

func (schemaResolver) Configuration() *configurationCascadeResolver {
	return &configurationCascadeResolver{}
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
	Value    *jsonString
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

	var value interface{}
	if args.Input.Value != nil {
		if err := json.Unmarshal([]byte(*args.Input.Value), &value); err != nil {
			return nil, err
		}
	}

	keyPath := jsonx.PropertyPath(args.Input.Property)
	_, err = r.doUpdateConfiguration(ctx, func(oldConfig string) (edits []jsonx.Edit, err error) {
		if args.Input.Value == nil {
			edits, _, err = jsonx.ComputePropertyRemoval(config, keyPath, formatOptions)
		} else {
			edits, _, err = jsonx.ComputePropertyEdit(config, keyPath, value, nil, formatOptions)
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
	actor := actor.FromContext(ctx)
	updatedSettings, err := store.Settings.CreateIfUpToDate(ctx, r.subject.toSubject(), r.input.LastID, actor.UID, newConfig)
	if err != nil {
		return 0, err
	}
	return updatedSettings.ID, nil
}

func (r *configurationMutationResolver) getCurrentConfig(ctx context.Context) (string, error) {
	// Get the settings file whose contents to mutate.
	settings, err := store.Settings.GetLatest(ctx, r.subject.toSubject())
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

var formatOptions = jsonx.FormatOptions{InsertSpaces: true, TabSize: 2, EOL: "\n"}
