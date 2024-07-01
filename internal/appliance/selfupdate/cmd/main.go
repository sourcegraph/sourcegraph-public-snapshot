package main

import (
	"github.com/sourcegraph/sourcegraph/appliance/selfupdate/k8s"
	"github.com/sourcegraph/sourcegraph/appliance/selfupdate/k8s/watcher"
	"github.com/sourcegraph/sourcegraph/appliance/selfupdate/server"
)

func main() {
	w := watcher.New()
	w.Run()
	server.New(k8s.New(w)).Run()
}
