// Package grpc is a set of shared code for implementing gRPC.
package grpc

import (
	"context"
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/featureflag"
)

const envGRPCEnabled = "SG_FEATURE_FLAG_GRPC"

func IsGRPCEnabled(ctx context.Context) bool {
	if val, err := strconv.ParseBool(os.Getenv(envGRPCEnabled)); err == nil {
		return val
	}
	return featureflag.FromContext(ctx).GetBoolOr("grpc", false)
}
