package main

import (
	"context"
	"log"
	"strconv"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	interval, _ = strconv.Atoi(env.Get("REPO_LIST_UPDATE_INTERVAL", "", "interval (in minutes) for checking code hosts (e.g. gitolite) for new repositories"))
)

func main() {
	if interval == 0 {
		log.Println("REPO_LIST_UPDATE_INTERVAL not set, not doing any updates")
		time.Sleep(10000 * 24 * time.Hour) // block forever and do nothing
	}

	for {
		time.Sleep(time.Duration(interval) * time.Minute)

		err := sourcegraph.InternalClient.GitoliteUpdateRepos(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("updated Gitolite repos")
	}
}
