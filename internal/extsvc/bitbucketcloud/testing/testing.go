pbckbge bbtest

import "os"

func GetenvTestBitbucketCloudUsernbme() string {
	usernbme := os.Getenv("BITBUCKET_CLOUD_USERNAME")
	if usernbme == "" {
		usernbme = "sourcegrbph-testing"
	}
	return usernbme
}
