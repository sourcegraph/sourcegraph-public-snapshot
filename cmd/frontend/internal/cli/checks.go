package cli

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/check"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

var checkCanReachGitserver = check.Check{
	Name:     "check_can_reach_gitserver",
	Interval: time.Minute,
	Run: func(ctx context.Context) (any, error) {

		type resp struct {
			Addr string `json:"addr"`
			Out  string `json:"out"`
		}

		addrs := conf.Get().ServiceConnections().GitServers
		resps := make([]resp, 0, len(addrs))

		checkAddr := func(addr string) ([]byte, error) {
			req, err := http.NewRequestWithContext(ctx, "GET", "http://"+addr+"/ping", nil)
			if err != nil {
				return nil, err
			}

			r, err := httpcli.InternalDoer.Do(req)
			if err != nil {
				return nil, err
			}

			if r.StatusCode != 200 {
				return nil, errors.New(r.Status)
			}

			b, err := io.ReadAll(r.Body)
			if err != nil {
				return nil, err
			}
			return b, nil
		}

		fail := false
		for _, addr := range addrs {
			b, err := checkAddr(addr)
			if err != nil {
				fail = true
				resps = append(resps, resp{
					Addr: addr,
					Out:  err.Error(),
				})
			} else {
				resps = append(resps, resp{
					Addr: addr,
					Out:  string(b),
				})
			}
		}

		var err error
		if fail {
			err = errors.New("could not reach one or more gitservers")
		}
		return resps, err
	},
}

var checkDummy = check.Check{
	Name:     "check_dummy",
	Interval: time.Second * 10,
	Run: func(ctx context.Context) (any, error) {
		return struct {
			Time time.Time
			Msg  string
		}{Time: time.Now(), Msg: "hello"}, nil
	},
}
