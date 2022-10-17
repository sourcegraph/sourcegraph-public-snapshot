package main

import (
	"context"
	"os"
	"strconv"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/audit"
)

func main() {
	args := os.Args

	// binary name, audit logs count
	if len(args) != 2 {
		os.Exit(-1)
	}

	ctx := context.Background()

	callbacks := log.Init(log.Resource{
		Name:       "Audit Resource",
		Namespace:  "Audit Integration Testing",
		Version:    "",
		InstanceID: "",
	})

	defer callbacks.Sync()

	logger := log.Scoped("test", "logger with sampling config")

	logsCount, err := strconv.Atoi(os.Args[1])
	if err != nil {
		os.Exit(-1)
	}

	for i := 0; i < logsCount; i++ {
		audit.Log(ctx, logger, audit.Record{
			Entity: "integration test",
			Action: "sampling testing",
			Fields: nil,
		})
	}
}
