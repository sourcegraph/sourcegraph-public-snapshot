package grpc

import (
	"context"
	"testing"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStreamPanicCatcher(t *testing.T) {
	logger, getLogs := logtest.Captured(t)

	streamInterceptor := NewStreamPanicCatcher(logger)
	err := streamInterceptor(
		nil,
		nil,
		&grpc.StreamServerInfo{FullMethod: "testmethod"},
		func(_ interface{}, _ grpc.ServerStream) error {
			panic("ouch")
		},
	)
	require.Error(t, err)
	require.Equal(t, codes.Internal, status.Code(err))

	logs := getLogs()
	require.Len(t, logs, 1)
	require.Equal(t, log.LevelError, logs[0].Level)
}

func TestUnaryPanicCatcher(t *testing.T) {
	logger, getLogs := logtest.Captured(t)

	unaryInterceptor := NewUnaryPanicCatcher(logger)
	_, err := unaryInterceptor(
		nil,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "testmethod"},
		func(_ context.Context, _ interface{}) (interface{}, error) {
			panic("ouch")
		},
	)
	require.Error(t, err)
	require.Equal(t, codes.Internal, status.Code(err))

	logs := getLogs()
	require.Len(t, logs, 1)
	require.Equal(t, log.LevelError, logs[0].Level)
}
