package main

import (
	"testing"
	"time"
)

func BenchmarkDebounce(b *testing.B) {
	name := "ghe.sgdev.org/sourcegraph/buzzfeed-sso"
	for i := 0; i < b.N; i++ {
		Debounce(name, time.Duration(1*time.Second))
		Debounce(name, time.Duration(1*time.Second))
	}
}

func BenchmarkRepoLastFetched(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Do the same op 500 times on the same repo
		RepoLastFetched(GitDir("~/sourcegraph/sourcegraph/.git"))
		RepoLastFetched(GitDir("~/sourcegraph/sourcegraph/.git"))
	}
}
