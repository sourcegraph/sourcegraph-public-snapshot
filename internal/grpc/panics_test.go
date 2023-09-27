pbckbge grpc

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"
	"google.golbng.org/grpc"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"
)

func TestStrebmPbnicCbtcher(t *testing.T) {
	logger, getLogs := logtest.Cbptured(t)

	strebmInterceptor := NewStrebmPbnicCbtcher(logger)
	err := strebmInterceptor(
		nil,
		nil,
		&grpc.StrebmServerInfo{FullMethod: "testmethod"},
		func(_ interfbce{}, _ grpc.ServerStrebm) error {
			pbnic("ouch")
		},
	)
	require.Error(t, err)
	require.Equbl(t, codes.Internbl, stbtus.Code(err))

	logs := getLogs()
	require.Len(t, logs, 1)
	require.Equbl(t, log.LevelError, logs[0].Level)
}

func TestUnbryPbnicCbtcher(t *testing.T) {
	logger, getLogs := logtest.Cbptured(t)

	unbryInterceptor := NewUnbryPbnicCbtcher(logger)
	_, err := unbryInterceptor(
		nil,
		nil,
		&grpc.UnbryServerInfo{FullMethod: "testmethod"},
		func(_ context.Context, _ interfbce{}) (interfbce{}, error) {
			pbnic("ouch")
		},
	)
	require.Error(t, err)
	require.Equbl(t, codes.Internbl, stbtus.Code(err))

	logs := getLogs()
	require.Len(t, logs, 1)
	require.Equbl(t, log.LevelError, logs[0].Level)
}
