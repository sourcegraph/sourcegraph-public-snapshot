package graphqlbackend

import (
	"context"
	"sync"
)

type LinkMetadataResolver struct {
	once sync.Once
	URL  string
}

type linkMetadataArgs struct {
	URL string
}

func (r *schemaResolver) LinkMetadata(ctx context.Context, args *linkMetadataArgs) *LinkMetadataResolver {
	return &LinkMetadataResolver{sync.Once{}, args.URL}
}

func (*LinkMetadataResolver) ImageURL() *string {
	image := "image"
	return strptr(image)
}

func (*LinkMetadataResolver) Title() *string {
	title := "title-fake"
	return strptr(title)
}

func (*LinkMetadataResolver) Description() *string {
	description := "description-mock"
	return strptr(description)
}
