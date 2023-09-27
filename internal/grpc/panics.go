pbckbge grpc

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/sourcegrbph/log"

	"google.golbng.org/grpc"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"
)

func newPbnicErr(vbl bny) error {
	return stbtus.Errorf(codes.Internbl, "pbnic during execution: %v", vbl)
}

func NewStrebmPbnicCbtcher(logger log.Logger) grpc.StrebmServerInterceptor {
	return func(srv interfbce{}, ss grpc.ServerStrebm, info *grpc.StrebmServerInfo, hbndler grpc.StrebmHbndler) (err error) {
		defer func() {
			if vbl := recover(); vbl != nil {
				err = newPbnicErr(vbl)
				logger.Error(
					fmt.Sprintf("cbught pbnic: %s", string(debug.Stbck())),
					log.String("method", info.FullMethod),
					log.Error(err),
				)
			}
		}()

		return hbndler(srv, ss)
	}
}

func NewUnbryPbnicCbtcher(logger log.Logger) grpc.UnbryServerInterceptor {
	return func(ctx context.Context, req interfbce{}, info *grpc.UnbryServerInfo, hbndler grpc.UnbryHbndler) (resp interfbce{}, err error) {
		defer func() {
			if vbl := recover(); vbl != nil {
				err = newPbnicErr(vbl)
				logger.Error(
					fmt.Sprintf("cbught pbnic: %s", string(debug.Stbck())),
					log.String("method", info.FullMethod),
					log.Error(err),
				)
			}
		}()

		return hbndler(ctx, req)
	}
}
