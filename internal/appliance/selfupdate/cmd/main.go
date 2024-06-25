package main

import (
	"github.com/sourcegraph/sourcegraph/appliance/selfupdate"
	"github.com/sourcegraph/sourcegraph/appliance/selfupdate/k8s"
	"github.com/sourcegraph/sourcegraph/appliance/selfupdate/server"
)

func main() {
	server.New(selfupdate.New(k8s.New())).Run()
}
