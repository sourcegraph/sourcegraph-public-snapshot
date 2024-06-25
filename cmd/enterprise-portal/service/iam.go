package service

import (
	"context"
	_ "embed"

	"github.com/redis/go-redis/v9"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/iam"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

//go:embed iam_model.fga
var iamModelDSL string

func newIAMClient(ctx context.Context, logger log.Logger, contract runtime.Contract, redisClient *redis.Client) (_ *iam.ClientV1, close func(), _ error) {
	iamClient, closeIAMClient, err := iam.NewClientV1(
		ctx,
		logger,
		contract,
		redisClient,
		iam.ClientV1Config{
			StoreName:             "enterprise-portal",
			AuthorizationModelDSL: iamModelDSL,
		},
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "new client")
	}
	return iamClient, closeIAMClient, nil
}
