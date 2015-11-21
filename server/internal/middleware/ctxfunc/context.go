package ctxfunc

import (
	"strings"

	"google.golang.org/grpc"
	"src.sourcegraph.com/sourcegraph/errcode"
)

func wrapErr(err error) error {
	if err == nil {
		return nil
	}

	// Don't double-wrap errors that are already gRPC errors.
	if strings.HasPrefix(err.Error(), "rpc error: code = ") {
		return err
	}

	return grpc.Errorf(errcode.GRPC(err), "%s", err.Error())
}
