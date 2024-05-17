package connectutil

import (
	"errors"

	"connectrpc.com/connect"

	"github.com/sourcegraph/log"
)

// InternalError logs an error and returns a connect error with a safe message.
// The logger should have trace context attached already.
func InternalError(logger log.Logger, err error, safeMsg string) error {
	logger.
		AddCallerSkip(1).
		Error(safeMsg,
			log.String("code", connect.CodeInternal.String()),
			log.Error(err))
	return connect.NewError(connect.CodeInternal, errors.New(safeMsg))
}
