//go:build !windows

package main

import (
	platform "golang.org/x/sys/unix"
)

const PLATFORM_SIGTERM = platform.SIGTERM
