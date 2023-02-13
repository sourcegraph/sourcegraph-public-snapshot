package graphql

import "context"

type Service interface {
	Foo(ctx context.Context) error
}
