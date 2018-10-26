// +build docker

// Package docker exists solely to ensure "github.com/sourcegraph/sourcegraph/cmd/server" exists as a dependency
package main

import (
	_ "github.com/sourcegraph/sourcegraph/cmd/server"
)
