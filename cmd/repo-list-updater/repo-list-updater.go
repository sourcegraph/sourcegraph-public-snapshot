package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	interval         = conf.Get().RepoListUpdateInterval
	frontendInternal = env.Get("SRC_FRONTEND_INTERNAL", "sourcegraph-frontend-internal", "HTTP address for internal frontend HTTP API.")
)

func main() {
	if interval == 0 {
		log.Println("REPO_LIST_UPDATE_INTERVAL not set, not doing any updates")
		time.Sleep(10000 * 24 * time.Hour) // block forever and do nothing
	}

	for {
		time.Sleep(time.Duration(interval) * time.Minute)

		resp, err := http.Post(fmt.Sprintf("http://%s/.api/repos-update", frontendInternal), "", nil)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("status: %s", resp.Status)
		resp.Body.Close()
	}
}
