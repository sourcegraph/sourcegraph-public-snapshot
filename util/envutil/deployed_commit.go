package envutil

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
)

var (
	// GitCommitID is the git commit ID of the deployed application,
	// if this application is deployed. If running locally, it is
	// empty.
	GitCommitID = readDeployedGitCommitID()
)

func readDeployedGitCommitID() string {
	b, err := ioutil.ReadFile(".git-commit-id")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("warn: Reading deployed git commit ID failed: %s", err)
		}
		return ""
	}
	return string(bytes.TrimSpace(b))
}
