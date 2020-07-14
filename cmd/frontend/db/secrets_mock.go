package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type MockSecrets struct {
	Delete                  func(ctx context.Context, id int32) error
	DeleteByKeyName         func(ctx context.Context, keyName string) error
	DeleteBySourceTypeAndID func(ctx context.Context, sourceType string, sourceID int32) error
	Get                     func(ctx context.Context, id int32) (*types.Secret, error)
	GetByKeyName            func(ctx context.Context, keyName string) (*types.Secret, error)
	GetBySourceTypeAndID    func(ctx context.Context, sourceType string, sourceID int32) (*types.Secret, error)
	Update                  func(ctx context.Context, id int32) error
	UpdateByKeyname         func(ctx context.Context, keyName string) error
	UpdateBySourceTypeAndID func(ctx context.Context, sourceType string, sourceID int32) error
}
