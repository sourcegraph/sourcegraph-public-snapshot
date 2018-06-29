package graphqlbackend

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func registryExtensionByID(ctx context.Context, id graphql.ID) (*registryExtensionMultiResolver, error) {
	registryExtensionID, err := unmarshalRegistryExtensionID(id)
	if err != nil {
		return nil, err
	}
	switch {
	case registryExtensionID.LocalID != 0:
		x, err := registryExtensionByIDInt32(ctx, registryExtensionID.LocalID)
		if err != nil {
			return nil, err
		}
		return &registryExtensionMultiResolver{local: x}, nil
	case registryExtensionID.RemoteID != nil:
		x, err := backend.GetRemoteRegistryExtension(ctx, "uuid", registryExtensionID.RemoteID.UUID)
		if err != nil {
			return nil, err
		}
		return &registryExtensionMultiResolver{remote: &registryExtensionRemoteResolver{x}}, nil
	default:
		return nil, errors.New("invalid registry extension ID")
	}
}

func registryExtensionByIDInt32(ctx context.Context, id int32) (*registryExtensionDBResolver, error) {
	if err := backend.CheckActorHasPlatformEnabled(ctx); err != nil {
		return nil, err
	}
	x, err := db.RegistryExtensions.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := backend.PrefixLocalExtensionID(x); err != nil {
		return nil, err
	}
	return &registryExtensionDBResolver{v: x}, nil
}

// registryExtensionID identifies a registry extension, either locally or on a remote
// registry. Exactly 1 field must be set.
type registryExtensionID struct {
	LocalID  int32                      `json:"l,omitempty"`
	RemoteID *registryExtensionRemoteID `json:"r,omitempty"`
}

func marshalRegistryExtensionID(id registryExtensionID) graphql.ID {
	return relay.MarshalID("RegistryExtension", id)
}

func unmarshalRegistryExtensionID(id graphql.ID) (registryExtensionID registryExtensionID, err error) {
	err = relay.UnmarshalSpec(id, &registryExtensionID)
	return
}

func readRegistryExtensionEnablement(extensionID, data string) *bool {
	var settings schema.Settings
	if err := conf.UnmarshalJSON(data, &settings); err != nil {
		// Don't treat as fatal because then we'd fail if any user has invalid settings JSON.
		return nil
	}
	if settings.Extensions == nil {
		return nil
	}
	extensionSettings, ok := settings.Extensions[extensionID]
	if !ok {
		return nil
	}
	v := !extensionSettings.Disabled
	return &v
}

type registryExtensionExtensionConfigurationSubjectsConnectionArgs struct {
	Users bool
	connectionArgs
}

func listExtensionConfigurationSubjects(ctx context.Context, extensionID string, args *registryExtensionExtensionConfigurationSubjectsConnectionArgs) (*extensionConfigurationSubjectConnection, error) {
	allSettings, err := db.Settings.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	var subjects []api.ConfigurationSubject
	for _, s := range allSettings {
		if !args.Users && s.Subject.User != nil {
			continue // exclude users
		}

		if v := readRegistryExtensionEnablement(extensionID, s.Contents); v != nil {
			subjects = append(subjects, s.Subject)
		}
	}
	return &extensionConfigurationSubjectConnection{
		connectionArgs: args.connectionArgs,
		subjects:       subjects,
	}, nil
}

func listRegistryExtensionUsers(ctx context.Context, extensionID string, args *connectionArgs) (*userConnectionResolver, error) {
	allSettings, err := db.Settings.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	// Build a model of the configuration, before cascading.
	site := false
	users := map[int32]bool{}
	orgs := map[int32]bool{}
	for _, s := range allSettings {
		switch {
		case s.Subject.Site:
			if v := readRegistryExtensionEnablement(extensionID, s.Contents); v != nil {
				site = *v
			}
		case s.Subject.Org != nil:
			if v := readRegistryExtensionEnablement(extensionID, s.Contents); v != nil {
				orgs[*s.Subject.Org] = *v
			}
		case s.Subject.User != nil:
			if v := readRegistryExtensionEnablement(extensionID, s.Contents); v != nil {
				users[*s.Subject.User] = *v
			}
		}
	}

	// Cascade to produce a final enablement value for each user.
	//
	cascadedUsers := map[int32]bool{}
	if site {
		allUsers, err := db.Users.List(ctx, nil)
		if err != nil {
			return nil, err
		}
		for _, user := range allUsers {
			cascadedUsers[user.ID] = true
		}
	}
	// Sort orgs for determinism, because multiple orgs' settings can apply to a user if they are in
	// multiple orgs.
	orgIDs := make([]int, 0, len(orgs))
	for orgID := range orgs {
		orgIDs = append(orgIDs, int(orgID))
	}
	sort.Ints(orgIDs)
	for _, orgID := range orgIDs {
		members, err := db.OrgMembers.GetByOrgID(ctx, int32(orgID))
		if err != nil {
			return nil, err
		}
		for _, m := range members {
			cascadedUsers[m.UserID] = orgs[int32(orgID)]
		}
	}
	for userID, v := range users {
		cascadedUsers[userID] = v
	}

	// Get all user IDs of users for whom the extension is enabled (after all cascading has been
	// performed).
	userIDs := make([]int32, 0, len(cascadedUsers))
	for userID, v := range cascadedUsers {
		if v {
			userIDs = append(userIDs, userID)
		}
	}

	// Return a UserConnection that fetches the set of user IDs we found.
	opt := db.UsersListOptions{UserIDs: userIDs}
	args.set(&opt.LimitOffset)
	return &userConnectionResolver{opt: opt}, nil
}

func viewerHasEnabledRegistryExtension(ctx context.Context, extensionID string) (bool, error) {
	// Consult the user's merged settings.
	merged, err := viewerMergedConfiguration(ctx)
	if err != nil {
		return false, err
	}
	var settings schema.Settings
	if err := json.Unmarshal([]byte(merged.Contents()), &settings); err != nil {
		return false, err
	}

	if settings.Extensions == nil {
		return false, nil
	}

	extensionSettings, ok := settings.Extensions[extensionID]
	return ok && !extensionSettings.Disabled, nil
}

func viewerCanConfigureRegistryExtension(ctx context.Context) (bool, error) {
	// Any authenticated user can use any extension.
	currentUser, err := currentUser(ctx)
	return currentUser != nil, err
}

func configuredExtensionFromRegistryExtension(ctx context.Context, extensionID string, args configuredExtensionFromRegistryExtensionArgs) (*configuredExtensionResolver, error) {
	var subject *configurationSubject
	if args.Subject != nil {
		// 🚨 SECURITY: The configurationSubjectByID func checks that the viewer can view the subject's
		// settings.
		var err error
		subject, err = configurationSubjectByID(ctx, *args.Subject)
		if err != nil {
			return nil, err
		}
	}
	return &configuredExtensionResolver{extensionID: extensionID, subject: subject}, nil
}
