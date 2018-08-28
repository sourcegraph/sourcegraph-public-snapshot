package graphqlbackend

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

type extensionRegistryMutationResult struct {
	id int32 // only used for local extensions
}

func (r *extensionRegistryMutationResult) Extension(ctx context.Context) (*registryExtensionDBResolver, error) {
	return registryExtensionByIDInt32(ctx, r.id)
}

func (*extensionRegistryResolver) CreateExtension(ctx context.Context, args *struct {
	Publisher graphql.ID
	Name      string
}) (*extensionRegistryMutationResult, error) {
	publisher, err := unmarshalRegistryPublisherID(args.Publisher)
	if err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Check that the current user can create an extension for this publisher.
	if err := publisher.viewerCanAdminister(ctx); err != nil {
		return nil, err
	}

	// Create the extension.
	id, err := db.RegistryExtensions.Create(ctx, publisher.userID, publisher.orgID, args.Name)
	if err != nil {
		return nil, err
	}
	return &extensionRegistryMutationResult{id: id}, nil
}

func (*extensionRegistryResolver) viewerCanAdministerExtension(ctx context.Context, id registryExtensionID) error {
	if id.LocalID == 0 {
		return errors.New("unable to administer extension on remote registry")
	}
	extension, err := db.RegistryExtensions.GetByID(ctx, id.LocalID)
	if err != nil {
		return err
	}
	return toRegistryPublisherID(extension).viewerCanAdminister(ctx)
}

func (*extensionRegistryResolver) UpdateExtension(ctx context.Context, args *struct {
	Extension graphql.ID
	Name      *string
}) (*extensionRegistryMutationResult, error) {
	id, err := unmarshalRegistryExtensionID(args.Extension)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check that the current user is authorized to update the extension.
	if err := (&extensionRegistryResolver{}).viewerCanAdministerExtension(ctx, id); err != nil {
		return nil, err
	}

	if err := db.RegistryExtensions.Update(ctx, id.LocalID, args.Name); err != nil {
		return nil, err
	}
	return &extensionRegistryMutationResult{id: id.LocalID}, nil
}

func (*extensionRegistryResolver) DeleteExtension(ctx context.Context, args *struct {
	Extension graphql.ID
}) (*EmptyResponse, error) {
	id, err := unmarshalRegistryExtensionID(args.Extension)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check that the current user is authorized to delete the extension.
	if err := (&extensionRegistryResolver{}).viewerCanAdministerExtension(ctx, id); err != nil {
		return nil, err
	}

	if err := db.RegistryExtensions.Delete(ctx, id.LocalID); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

func (*extensionRegistryResolver) PublishExtension(ctx context.Context, args *struct {
	ExtensionID string
	Manifest    string
	Bundle      *string
	SourceMap   *string
	Force       bool
}) (*extensionRegistryMutationResult, error) {
	// Add the prefix if needed, for ease of use.
	configuredPrefix := backend.GetLocalRegistryExtensionIDPrefix()
	prefix, _, _, err := backend.SplitExtensionID(args.ExtensionID)
	if err != nil {
		return nil, err
	}
	if prefix == "" && configuredPrefix != nil {
		args.ExtensionID = *configuredPrefix + "/" + args.ExtensionID
	}

	prefix, _, isLocal, err := backend.ParseExtensionID(args.ExtensionID)
	if err != nil {
		return nil, err
	}
	if !isLocal {
		return nil, fmt.Errorf("unable to publish remote extension %q (publish it directly to the registry on %q)", args.ExtensionID, prefix)
	}

	// Get or create the extension to publish.
	localExtension, remoteExtension, err := backend.GetExtensionByExtensionID(ctx, args.ExtensionID)
	// Check that the extension doesn't refer to an existing remote extension. This can be true even
	// with the !isLocal check above in the case of synthesized BACKCOMPAT extensions.
	if remoteExtension != nil {
		return nil, fmt.Errorf("unable to publish extension %q because it conflicts with an existing non-local extension", args.ExtensionID)
	}
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}

	// Create the extension if needed.
	var id registryExtensionID
	if localExtension == nil {
		_, publisherName, extensionName, err := backend.SplitExtensionID(args.ExtensionID)
		if err != nil {
			return nil, err
		}
		publisher, err := db.RegistryExtensions.GetPublisher(ctx, publisherName)
		if err != nil {
			return nil, err
		}
		publisherID := registryPublisherID{userID: publisher.UserID, orgID: publisher.OrgID}
		// 🚨 SECURITY: Check that the current user can create an extension for this publisher.
		if err := publisherID.viewerCanAdminister(ctx); err != nil {
			return nil, err
		}

		// Create the extension.
		xid, err := db.RegistryExtensions.Create(ctx, publisherID.userID, publisherID.orgID, extensionName)
		if err != nil {
			return nil, err
		}
		id.LocalID = xid
	} else {
		id.LocalID = localExtension.ID
	}

	// 🚨 SECURITY: Check that the current user is authorized to publish the extension.
	if err := (&extensionRegistryResolver{}).viewerCanAdministerExtension(ctx, id); err != nil {
		return nil, err
	}

	// Validate the manifest.
	if err := backend.ValidateExtensionManifest(args.Manifest); err != nil {
		if !args.Force {
			return nil, fmt.Errorf("invalid extension manifest: %s", err)
		}
	}

	release := db.RegistryExtensionRelease{
		RegistryExtensionID: id.LocalID,
		CreatorUserID:       actor.FromContext(ctx).UID,
		ReleaseTag:          "release",
		Manifest:            args.Manifest,
		Bundle:              args.Bundle,
	}
	if _, err := db.RegistryExtensionReleases.Create(ctx, &release); err != nil {
		return nil, err
	}
	return &extensionRegistryMutationResult{id: id.LocalID}, nil
}
