package cli

import (
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
	Run: func() any {

		type resp struct {
			Addr string `json:"addr"`
			Ok   bool   `json:"ok"`
			Msg  string `json:"msg"`
		}

		addrs := conf.Get().ServiceConnections().GitServers
		resps := make([]resp, 0, len(addrs))

		checkAddr := func(addr string) ([]byte, error) {
			req, err := http.NewRequest("GET", "http://"+addr+"/ping", nil)
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

		for _, addr := range addrs {
			b, err := checkAddr(addr)
			if err != nil {
				resps = append(resps, resp{
					Addr: addr,
					Ok:   false,
					Msg:  err.Error(),
				})
			} else {
				resps = append(resps, resp{
					Addr: addr,
					Ok:   true,
					Msg:  string(b),
				})
			}
		}

		return resps
	},
}
