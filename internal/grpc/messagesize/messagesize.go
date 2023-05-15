package messagesize

import (
	"fmt"
	"math"
	"os"

	"google.golang.org/grpc"

	"github.com/dustin/go-humanize"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	envClientMessageSize = "SRC_GRPC_CLIENT_MAX_MESSAGE_SIZE"
	envServerMessageSize = "SRC_GRPC_SERVER_MAX_MESSAGE_SIZE"

	smallestAllowedMaxMessageSize = 4 * 1024 * 1024 // 4 MB: There isn't a scenario where we'd want to dip below the default of 4MB.
	largestAllowedMaxMessageSize  = math.MaxInt     //
)

// ClientMessageSizeFromEnv returns a grpc.DialOption that sets the maximum message size for gRPC clients.
func ClientMessageSizeFromEnv(l log.Logger) []grpc.DialOption {
	messageSize, err := getMessageSizeBytesFromEnv(envClientMessageSize)

	if err != nil {
		// Log only if the error is not an envNotSetError
		var e envNotSetError
		if !errors.As(err, &e) {
			l.Warn("failed to get gRPC client message size, setting to default value", log.Error(err), log.String("default", "4MB"))
		}

		return nil
	}

	return []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(messageSize),
			grpc.MaxCallSendMsgSize(messageSize),
		),
	}
}

// ServerMessageSizeFromEnv returns a grpc.ServerOption that sets the maximum message size for gRPC servers.
func ServerMessageSizeFromEnv(l log.Logger) []grpc.ServerOption {
	messageSize, err := getMessageSizeBytesFromEnv(envServerMessageSize)

	if err != nil {
		// Log only if the error is not an envNotSetError
		var e envNotSetError
		if !errors.As(err, &e) {
			l.Warn("failed to get gRPC server message size, setting to default value", log.Error(err), log.String("default", "4MB"))
		}

		return nil
	}

	return []grpc.ServerOption{
		grpc.MaxRecvMsgSize(messageSize),
		grpc.MaxSendMsgSize(messageSize),
	}
}

func getMessageSizeBytesFromEnv(envVar string) (size int, err error) {
	rawSize, set := os.LookupEnv(envVar)
	if !set {
		return 0, envNotSetError{envVar: envVar}
	}

	sizeBytes, err := humanize.ParseBytes(rawSize)
	if err != nil {
		return 0, errors.Wrapf(err, "getMessageSizeBytesFromEnv: parsing %q as bytes", rawSize)
	}

	if sizeBytes < smallestAllowedMaxMessageSize || int(sizeBytes) > largestAllowedMaxMessageSize {
		return 0, errors.Errorf("getMessageSizeBytesFromEnv: message size %d is outside of allowed range [%d, %d]", size, smallestAllowedMaxMessageSize, largestAllowedMaxMessageSize)
	}

	return int(sizeBytes), nil
}

// envNotSetError occurs when the environment variable specified by envVar is not set.
type envNotSetError struct {
	// envVar is the name of the environment variable that was not set
	envVar string
}

func (e envNotSetError) Error() string {
	return fmt.Sprintf("environment variable %q not set", e.envVar)
}
