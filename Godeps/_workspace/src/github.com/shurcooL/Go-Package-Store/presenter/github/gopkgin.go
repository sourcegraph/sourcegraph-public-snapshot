package github

import (
	"fmt"
	"strings"
)

func gopkgInImportPathToGitHub(gopkgInImportPath string) (gitHubOwner, gitHubRepo string, err error) {
	afterPrefix := gopkgInImportPath[len("gopkg.in/"):]
	importPathElements0 := strings.Split(afterPrefix, ".")
	if len(importPathElements0) != 2 {
		return "", "", fmt.Errorf("len(importPathElements0) != 2: %v", importPathElements0)
	}
	importPathElements1 := strings.Split(importPathElements0[0], "/")
	importPath := "github.com/"
	if len(importPathElements1) == 1 { // gopkg.in/pkg.v3 -> github.com/go-pkg/pkg
		importPath += "go-" + importPathElements1[0] + "/" + importPathElements1[0]
	} else if len(importPathElements1) == 2 { // gopkg.in/user/pkg.v3 -> github.com/user/pkg
		importPath += importPathElements1[0] + "/" + importPathElements1[1]
	} else {
		return "", "", fmt.Errorf("len(importPathElements1) != 1 nor 2: %v", importPathElements1)
	}
	importPathElements := strings.Split(importPath, "/")
	return importPathElements[1], importPathElements[2], nil
}
