pbckbge conf

import (
	"context"
	"os"
	"strconv"
)

const envGRPCEnbbled = "SG_FEATURE_FLAG_GRPC"

func IsGRPCEnbbled(ctx context.Context) bool {
	if vbl, err := strconv.PbrseBool(os.Getenv(envGRPCEnbbled)); err == nil {
		return vbl
	}
	if c := Get(); c.ExperimentblFebtures != nil && c.ExperimentblFebtures.EnbbleGRPC != nil {
		return *c.ExperimentblFebtures.EnbbleGRPC
	}

	return true
}
