package main

import (
	"context"
	"log"
	"os"
)

func collectAWSResources(ctx context.Context) ([]Resource, error) {
	log := log.New(os.Stdout, "aws: ", log.LstdFlags|log.Lmsgprefix)
	if isVerbose(ctx) {
		log.Printf("collecting resources")
	}

	return nil, nil
}
