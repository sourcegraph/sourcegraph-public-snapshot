// +build !enable_azul3d

package github

import "errors"

// azul3d presenter is temporarily disabled due to an upstream issue.
// It can be re-enabled once https://github.com/azul3d/website/issues/11 is resolved.
//
// It can be enabled by using "enable_azul3d" build tag.
func azul3dOrgImportPathToGitHub(azul3dOrgImportPath string) (gitHubOwner, gitHubRepo string, err error) {
	return "", "", errors.New("azul3d presenter disabled until https://github.com/azul3d/website/issues/11 is resolved")
}
