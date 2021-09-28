package main

import "fmt"

func makeRepoName(repoName string) string {
	return fmt.Sprintf("github.com/%s/%s", "sourcegraph-testing", repoName)
}
