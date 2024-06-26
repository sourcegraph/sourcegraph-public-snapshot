package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/appliance/shared"
	"github.com/sourcegraph/sourcegraph/internal/sanitycheck"
	"github.com/sourcegraph/sourcegraph/internal/service/svcmain"
)

func main() {
	sanitycheck.Pass()
	svcmain.SingleServiceMainWithoutConf(shared.Service, nil, svcmain.OutOfBandConfiguration{})
}
