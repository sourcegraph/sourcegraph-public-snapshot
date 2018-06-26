package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
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
	Manifest  *string
}) (*extensionRegistryMutationResult, error) {
	id, err := unmarshalRegistryExtensionID(args.Extension)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check that the current user is authorized to update the extension.
	if err := (&extensionRegistryResolver{}).viewerCanAdministerExtension(ctx, id); err != nil {
		return nil, err
	}

	if err := db.RegistryExtensions.Update(ctx, id.LocalID, args.Name, args.Manifest); err != nil {
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
