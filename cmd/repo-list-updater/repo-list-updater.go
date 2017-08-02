package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var interval, _ = strconv.Atoi(env.Get("REPO_LIST_UPDATE_INTERVAL", "", "interval (in minutes) for checking code hosts (e.g. gitolite) for new repositories"))

func main() {
	if interval == 0 {
		log.Println("REPO_LIST_UPDATE_INTERVAL not set, not doing any updates")
		select {} // block forever and do nothing
	}

	for {
		time.Sleep(time.Duration(interval) * time.Minute)

		resp, err := http.Post("http://sourcegraph-frontend-internal/.api/repos-update", "", nil)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("status: %s", resp.Status)
		resp.Body.Close()
	}
}
