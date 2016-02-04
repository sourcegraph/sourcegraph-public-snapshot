// +build enable_azul3d

package github

import (
	"errors"
	"net/url"
	"strings"

	"azul3d.org/semver.v2"
)

var azul3dOrgMatcher = semver.GitHub("azul3d")

func azul3dOrgImportPathToGitHub(azul3dOrgImportPath string) (gitHubOwner, gitHubRepo string, err error) {
	u, err := url.Parse("https://" + azul3dOrgImportPath)
	if err != nil {
		return "", "", err
	}

	repo, err := azul3dOrgMatcher.Match(u)
	if err != nil {
		return "", "", err
	}
	repo.URL.Path = strings.TrimSuffix(repo.URL.Path, ".git")

	pathElements := strings.Split(repo.URL.Path, "/")
	if len(pathElements) < 2 {
		return "", "", errors.New("azul3dOrgImportPathToGitHub: len(pathElements) < 2")
	}
	gitHubOwner, gitHubRepo = pathElements[0], pathElements[1]

	return gitHubOwner, gitHubRepo, err
}
