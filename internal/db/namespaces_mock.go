package db

import "context"

type MockNamespaces struct {
	GetByName func(ctx context.Context, name string) (*Namespace, error)
}
