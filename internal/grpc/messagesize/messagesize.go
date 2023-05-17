package messagesize

import (
	"fmt"
	"math"
	"os"
	"sync"

	"google.golang.org/grpc"

	"github.com/dustin/go-humanize"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	smallestAllowedMaxMessageSize = uint64(4 * 1024 * 1024) // 4 MB: There isn't a scenario where we'd want to dip below the default of 4MB.
	largestAllowedMaxMessageSize  = uint64(math.MaxInt)     // This is the largest allowed value for the type accepted by the grpc.MaxSize[...] options.

	logClientWarningOnce sync.Once // Ensures that we only log the warning once, since we might call ClientMessageSizeFromEnv multiple times.
	envClientMessageSize = "SRC_GRPC_CLIENT_MAX_MESSAGE_SIZE"

	logServerWarningOnce sync.Once // Ensures that we only log the warning once, since we might call ServerMessageSizeFromEnv multiple times.
	envServerMessageSize = "SRC_GRPC_SERVER_MAX_MESSAGE_SIZE"
)

// ClientMessageSizeFromEnv returns a slice of grpc.DialOptions that set the maximum message size for gRPC clients if
// the "SRC_GRPC_CLIENT_MAX_MESSAGE_SIZE" environment variable is set to a valid size value (ex: "40 MB").
//
// If the environment variable is not set or if the size value is invalid (too small, not parsable, etc.), it returns nil.
func ClientMessageSizeFromEnv(l log.Logger) []grpc.DialOption {
	messageSize, err := getMessageSizeBytesFromEnv(envClientMessageSize, smallestAllowedMaxMessageSize, largestAllowedMaxMessageSize)
	if err != nil {
		// Log only if the error is not an envNotSetError
		var e *envNotSetError
		if !errors.As(err, &e) {
			logClientWarningOnce.Do(func() {
				l.Warn("failed to get gRPC client message size, setting to default value",
					log.Error(err),
					log.String("default", humanize.IBytes(smallestAllowedMaxMessageSize)),
				)
			})
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

// ServerMessageSizeFromEnv returns a slice of grpc.ServerOption that set the maximum message size for gRPC servers if
// the "SRC_GRPC_SERVER_MAX_MESSAGE_SIZE" environment variable is set to a valid size value (ex: "40 MB").
//
// If the environment variable is not set or if the size value is invalid (too small, not parsable, etc.), it returns nil.
func ServerMessageSizeFromEnv(l log.Logger) []grpc.ServerOption {
	messageSize, err := getMessageSizeBytesFromEnv(envServerMessageSize, smallestAllowedMaxMessageSize, largestAllowedMaxMessageSize)
	if err != nil {
		// Log only if the error is not an envNotSetError
		var e *envNotSetError
		if !errors.As(err, &e) {
			logServerWarningOnce.Do(func() {
				l.Warn("failed to get gRPC server message size, using default value",
					log.Error(err),
					log.String("default", humanize.IBytes(smallestAllowedMaxMessageSize)),
				)
			})
		}

		return nil
	}

	return []grpc.ServerOption{
		grpc.MaxRecvMsgSize(messageSize),
		grpc.MaxSendMsgSize(messageSize),
	}
}

// getMessageSizeBytesFromEnv reads the environment variable specified by envVar
// and returns the message size in bytes within the range [minSize, maxSize].
//
// If the environment variable is not set or the value is outside the allowed range,
// it returns an error.
func getMessageSizeBytesFromEnv(envVar string, minSize, maxSize uint64) (size int, err error) {
	rawSize, set := os.LookupEnv(envVar)
	if !set {
		return 0, &envNotSetError{envVar: envVar}
	}

	sizeBytes, err := humanize.ParseBytes(rawSize)
	if err != nil {
		return 0, &parseError{
			rawSize: rawSize,
			err:     err,
		}
	}

	if sizeBytes < minSize || sizeBytes > maxSize {
		return 0, &sizeOutOfRangeError{
			size: humanize.IBytes(sizeBytes),
			min:  humanize.IBytes(minSize),
			max:  humanize.IBytes(maxSize),
		}
	}

	return int(sizeBytes), nil
}

// parseError occurs when the environment variable's value cannot be parsed as a byte size.
type parseError struct {
	// rawSize is the raw size string that was attempted to be parsed
	rawSize string
	// err is the error that occurred while parsing rawSize
	err error
}

func (e *parseError) Error() string {
	return fmt.Sprintf("failed to parse %q as bytes: %s", e.rawSize, e.err)
}

func (e *parseError) Unwrap() error {
	return e.err
}

// sizeOutOfRangeError occurs when the environment variable's value is outside of the allowed range.
type sizeOutOfRangeError struct {
	// size is the size that was out of range
	size string
	// min is the minimum allowed size
	min string
	// max is the maximum allowed size
	max string
}

func (e *sizeOutOfRangeError) Error() string {
	return fmt.Sprintf("size %s is outside of allowed range [%s, %s]", e.size, e.min, e.max)
}

// envNotSetError occurs when the environment variable specified by envVar is not set.
type envNotSetError struct {
	// envVar is the name of the environment variable that was not set
	envVar string
}

func (e *envNotSetError) Error() string {
	return fmt.Sprintf("environment variable %q not set", e.envVar)
}
