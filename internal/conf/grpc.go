package conf

import (
	"context"
	"os"
	"strconv"
)

const envGRPCEnabled = "SG_FEATURE_FLAG_GRPC"

func IsGRPCEnabled(ctx context.Context) bool {
	if val, err := strconv.ParseBool(os.Getenv(envGRPCEnabled)); err == nil {
		return val
	}
	if c := Get(); c.ExperimentalFeatures != nil && c.ExperimentalFeatures.EnableGRPC != nil {
		return *c.ExperimentalFeatures.EnableGRPC
	}

	return true
}
