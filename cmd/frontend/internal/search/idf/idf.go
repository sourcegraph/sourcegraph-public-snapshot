// Package idf computes and stores the inverse document frequency (IDF) of a set of repositories.
//
// TODO(beyang): should probably move this elsewhere
package idf

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)


var redisCache = rcache.NewWithTTL(redispool.Cache, "idf-index", 10*24*60*60)


func Update(ctx context.Context, repoID api.RepoID, value interface{}) error {
	fmt.Printf("# idf.Update(%v)", repoID)
	b, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "idf.Update")
	}
	redisCache.Set(fmt.Sprintf("repo:%v", repoID), b)
	return nil
}


func Get(ctx context.Context, repoID api.RepoID) ([]byte, error) {
	fmt.Printf("# idf.Get(%v)", repoID)
	b, ok := redisCache.Get(fmt.Sprintf("repo:%v", repoID))
	if !ok {
		return nil, fmt.Errorf("idf.Get: repo %s not found", string(repoID))
	}
	return b, nil
}
