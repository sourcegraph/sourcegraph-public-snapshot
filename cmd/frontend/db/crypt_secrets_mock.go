package db

import (
	"context"
	"testing"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type MockCryptSecrets struct {
	Delete func(ctx context.Context, id int32) error
	Update func(ctx context.Context, id int32) error
	Get func(ctx context.Context, id int32) (*types.CryptSecret, error)
	GetBySourceTypeAndID func(ctx context.Context, sourceType string, sourceID int32) (*types.CryptSecret, error)
	DeleteBySourceTypeAndID func(ctx context.Context, sourceType string, sourceID int32) error
	UpdateBySourceTypeAndID func(ctx context.Context, sourceType string, sourceID int32) error
}

func (s *MockCryptSecrets) MockGet(t *testing.T, crypt int32) (called *bool) {
	s.Get = func(ctx context.Context, mockID int32) (*types.CryptSecret, error) {
		*called = true
		if mockID != crypt {
			t.Errorf("Retrieved %d, but wanted %d", mockID, crypt)
			return nil, cryptNotFoundError{}
		}
		return &types.CryptSecret{ID: mockID}, nil
	}
	return
}

func (s *MockCryptSecrets) MockGetBySourceTypeAndID(t *testing.T, sourceType string, sourceID int32) (called *bool) {
	s.GetBySourceTypeAndID = func(ctx context.Context, mockSourceType string, mockID int32) (*types.CryptSecret, error) {
		*called = true
		if mockID != sourceID {
			t.Errorf("Retrieved %d, but wanted %d", mockID, sourceID)
			return nil, cryptNotFoundError{}
		}
		return &types.CryptSecret{ID: mockID}, nil
	}
	return
}
