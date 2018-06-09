package main

import (
	"log"
	"os"
)

// setDefaultEnv will set the environment variable if it is not set.
func setDefaultEnv(k, v string) string {
	if s, ok := os.LookupEnv(k); ok {
		return s
	}
	err := os.Setenv(k, v)
	if err != nil {
		log.Fatal(err)
	}
	return v
}
