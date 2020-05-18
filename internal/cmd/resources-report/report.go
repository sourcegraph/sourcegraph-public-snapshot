package main

import (
	"context"
	"fmt"
)

type Platform string

const (
	PlatformGCP Platform = "gcp"
	PlatformAWS Platform = "aws"
)

type Resource struct {
	Platform   Platform
	Identifier string
	Type       string
	Location   string
	Owner      string
	Meta       map[string]interface{}
}

func reportString(resources []Resource) string {
	var output string
	for _, r := range resources {
		output += fmt.Sprintf(" * %+v\n", r)
	}
	return output
}

func reportToSlack(ctx context.Context, webhook string, resources []Resource) error {
	return nil
}
