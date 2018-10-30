package httpapi

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/httpapi"
	"github.com/sourcegraph/sourcegraph/xlang"
)

func init() {
	httpapi.XLangNewClient = func() (httpapi.XLangClient, error) {
		c, err := xlang.UnsafeNewDefaultClient()
		if err != nil {
			return nil, err
		}
		return &xclient{Client: c}, nil
	}
}
