package typesystem

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/karlseguin/ccache/v3"
	"github.com/oklog/ulid/v2"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"golang.org/x/sync/singleflight"

	"github.com/openfga/openfga/pkg/storage"
)

const (
	typesystemCacheTTL = 168 * time.Hour // 7 days.
)

// TypesystemResolverFunc is a function type that implementations
// can use to provide lookup and resolution of a Typesystem.
type TypesystemResolverFunc func(ctx context.Context, storeID, modelID string) (*TypeSystem, error)

// MemoizedTypesystemResolverFunc returns a TypesystemResolverFunc that fetches the provided authorization
// model (if provided) or looks up the latest authorization model. It then constructs a TypeSystem from
// the resolved model, and memoizes the type-system resolution. If another lookup of the same model occurs,
// the earlier constructed TypeSystem will be used.
//
// The memoized resolver function is designed for concurrent use.
func MemoizedTypesystemResolverFunc(datastore storage.AuthorizationModelReadBackend) (TypesystemResolverFunc, func()) {
	lookupGroup := singleflight.Group{}

	cache := ccache.New(ccache.Configure[*TypeSystem]())

	return func(ctx context.Context, storeID, modelID string) (*TypeSystem, error) {
		ctx, span := tracer.Start(ctx, "MemoizedTypesystemResolverFunc")
		defer span.End()

		var err error

		if modelID != "" {
			if _, err := ulid.Parse(modelID); err != nil {
				return nil, ErrModelNotFound
			}
		}

		var v interface{}
		var key string
		if modelID == "" {
			v, err, _ = lookupGroup.Do(fmt.Sprintf("FindLatestAuthorizationModel:%s", storeID), func() (interface{}, error) {
				return datastore.FindLatestAuthorizationModel(ctx, storeID)
			})
			if err != nil {
				if errors.Is(err, storage.ErrNotFound) {
					return nil, ErrModelNotFound
				}

				return nil, fmt.Errorf("failed to FindLatestAuthorizationModel: %w", err)
			}

			model := v.(*openfgav1.AuthorizationModel)
			key = fmt.Sprintf("%s/%s", storeID, model.GetId())
		} else {
			key = fmt.Sprintf("%s/%s", storeID, modelID)
			item := cache.Get(key)
			if item != nil {
				return item.Value(), nil
			}

			v, err, _ = lookupGroup.Do(fmt.Sprintf("ReadAuthorizationModel:%s/%s", storeID, modelID), func() (interface{}, error) {
				return datastore.ReadAuthorizationModel(ctx, storeID, modelID)
			})
			if err != nil {
				if errors.Is(err, storage.ErrNotFound) {
					return nil, ErrModelNotFound
				}

				return nil, fmt.Errorf("failed to ReadAuthorizationModel: %w", err)
			}
		}

		model := v.(*openfgav1.AuthorizationModel)

		typesys, err := NewAndValidate(ctx, model)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidModel, err)
		}

		cache.Set(key, typesys, typesystemCacheTTL)

		return typesys, nil
	}, cache.Stop
}
