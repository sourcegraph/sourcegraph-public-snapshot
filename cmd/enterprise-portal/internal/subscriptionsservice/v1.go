package subscriptionsservice

import (
	"net/http"

	"github.com/sourcegraph/log"

	subscriptionsv1connect "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1/v1connect"
)

const Name = subscriptionsv1connect.SubscriptionsServiceName

func RegisterV1(logger log.Logger, mux *http.ServeMux) {
	mux.Handle(subscriptionsv1connect.NewSubscriptionsServiceHandler(&handlerV1{
		logger: logger.Scoped("subscriptions.v1"),
	}))
}

type handlerV1 struct {
	subscriptionsv1connect.UnimplementedSubscriptionsServiceHandler

	logger log.Logger
}

var _ subscriptionsv1connect.SubscriptionsServiceHandler = (*handlerV1)(nil)
