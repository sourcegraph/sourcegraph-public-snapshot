package main

import (
	"errors"
	"syscall"
)

// This whole file probably needs work to handle things like being run on different OSes
// My understanding is that if `getrusage` is different for your machine, then you'll get
// different results.
// Something to consider for later. That's why the code lives in a separate place though.

func MaxMemoryInKB(usage interface{}) (int64, error) {
	sysUsage, ok := usage.(*syscall.Rusage)

	if !ok {
		return -1, errors.New("Could not convert usage")
	}

	return sysUsage.Maxrss, nil
}
