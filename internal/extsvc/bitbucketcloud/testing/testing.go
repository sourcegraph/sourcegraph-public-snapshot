package bbtest

import "os"

func GetenvTestBitbucketCloudUsername() string {
	username := os.Getenv("BITBUCKET_CLOUD_USERNAME")
	if username == "" {
		username = "sourcegraph-testing"
	}
	return username
}
