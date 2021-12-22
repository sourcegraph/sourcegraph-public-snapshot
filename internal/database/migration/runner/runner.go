package runner

import (
	"context"
)

type Runner struct {
	storeFactories map[string]StoreFactory
}

type StoreFactory func(ctx context.Context) (Store, error)

func NewRunner(storeFactories map[string]StoreFactory) *Runner {
	return &Runner{
		storeFactories: storeFactories,
	}
}
