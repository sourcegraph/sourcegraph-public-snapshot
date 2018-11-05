package graphqlbackend

import (
	"context"
	"errors"
)

type resource interface {
	ResourceURI() string
}

type resourceResolver struct {
	resource
}

// func (r *resourceResolver) ToXyz() (*accessTokenResolver, bool) {
// 	n, ok := r.node.(*accessTokenResolver)
// 	return n, ok
// }

func (r *schemaResolver) Resource(ctx context.Context, args *struct{ URI string }) (*resourceResolver, error) {
	resource, err := resourceByURI(ctx, args.URI)
	if err != nil {
		return nil, err
	}
	return &resourceResolver{resource}, nil
}

func resourceByURI(ctx context.Context, uri string) (resource, error) {
	return nil, errors.New("invalid URI")
}
