package messagesize

import (
	"fmt"
	"math"

	"google.golang.org/grpc"

	"github.com/dustin/go-humanize"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	smallestAllowedMaxMessageSize = uint64(4 * 1024 * 1024) // 4 MB: There isn't a scenario where we'd want to dip below the default of 4MB.
	largestAllowedMaxMessageSize  = uint64(math.MaxInt)     // This is the largest allowed value for the type accepted by the grpc.MaxSize[...] options.

	envClientMessageSize = env.Get("SRC_GRPC_CLIENT_MAX_MESSAGE_SIZE", messageSizeDisabled, fmt.Sprintf("set the maximum message size for gRPC clients (ex: %q)", "40MB"))
	envServerMessageSize = env.Get("SRC_GRPC_SERVER_MAX_MESSAGE_SIZE", messageSizeDisabled, fmt.Sprintf("set the maximum message size for gRPC servers (ex: %q)", "40MB"))

	messageSizeDisabled = "message_size_disabled" // sentinel value for when the message size env var isn't set
)

// MustGetClientMessageSizeFromEnv returns a slice of grpc.DialOptions that set the maximum message size for gRPC clients if
// the "SRC_GRPC_CLIENT_MAX_MESSAGE_SIZE" environment variable is set to a valid size value (ex: "40 MB").
//
// If the environment variable isn't set, it returns nil.
// If the size value in the environment variable is invalid (too small, not parsable, etc.), it panics.
func MustGetClientMessageSizeFromEnv() []grpc.DialOption {
	if envClientMessageSize == messageSizeDisabled {
		return nil
	}

	messageSize, err := getMessageSizeBytesFromString(envClientMessageSize, smallestAllowedMaxMessageSize, largestAllowedMaxMessageSize)
	if err != nil {
		panic(fmt.Sprintf("failed to get gRPC client message size: %s", err))
	}

	return []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(messageSize),
			grpc.MaxCallSendMsgSize(messageSize),
		),
	}
}

// MustGetServerMessageSizeFromEnv returns a slice of grpc.ServerOption that set the maximum message size for gRPC servers if
// the "SRC_GRPC_SERVER_MAX_MESSAGE_SIZE" environment variable is set to a valid size value (ex: "40 MB").
//
// If the environment variable isn't set, it returns nil.
// If the size value in the environment variable is invalid (too small, not parsable, etc.), it panics.
func MustGetServerMessageSizeFromEnv() []grpc.ServerOption {
	if envServerMessageSize == messageSizeDisabled {
		return nil
	}

	messageSize, err := getMessageSizeBytesFromString(envServerMessageSize, smallestAllowedMaxMessageSize, largestAllowedMaxMessageSize)
	if err != nil {
		panic(fmt.Sprintf("failed to get gRPC server message size: %s", err))
	}

	return []grpc.ServerOption{
		grpc.MaxRecvMsgSize(messageSize),
		grpc.MaxSendMsgSize(messageSize),
	}
}

// getMessageSizeBytesFromEnv parses rawSize returns the message size in bytes within the range [minSize, maxSize].
//
// If rawSize isn't a valid size is not set or the value is outside the allowed range, it returns an error.
func getMessageSizeBytesFromString(rawSize string, minSize, maxSize uint64) (size int, err error) {
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
