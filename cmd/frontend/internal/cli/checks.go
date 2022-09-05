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
	"github.com/sourcegraph/sourcegraph/internal/search"
)

var checkCanReachGitserver = check.Check{
	Name:     "check_can_reach_gitserver",
	Interval: time.Minute,
	Run: func(ctx context.Context) (any, error) {
		addrs := conf.Get().ServiceConnections().GitServers

		var addrs2 []string
		for _, addr := range addrs {
			addrs2 = append(addrs2, "http://"+addr+"/ping")
		}
		return checkCanReachHTTP(ctx, addrs2)
	},
}

var checkCanReachSearcher = check.Check{
	Name:     "check_can_reach_searcher",
	Interval: time.Minute,
	Run: func(ctx context.Context) (any, error) {
		addrs, err := search.SearcherURLs().Endpoints()
		if err != nil {
			return "", err
		}
		var addrs2 []string
		for _, addr := range addrs {
			addrs2 = append(addrs2, addr+"/healthz")
		}
		return checkCanReachHTTP(ctx, addrs2)
	},
}

func checkCanReachHTTP(ctx context.Context, addrs []string) (any, error) {
	type resp struct {
		Addr string `json:"addr"`
		Out  string `json:"out"`
	}

	resps := make([]resp, 0, len(addrs))

	checkAddr := func(addr string) ([]byte, error) {
		req, err := http.NewRequestWithContext(ctx, "GET", addr, nil)
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

	// Aggregate error.
	var err error
	if fail {
		err = errors.New("could not reach one or more servers")
	}
	return resps, err
}
