package plan

import (
	"fmt"
	"strings"
)

// Drone Docker images for running each toolchain. Remove the tag when
// developing to make it easier to test out changes to a given toolchain. E.g.,
// `droneSrclibGoImage = "sourcegraph/srclib-go"`.
var (
	droneSrclibBashImage       = "sourcegraph/srclib-bash:25395370b151f2a98a8430263323374b8c55bd0f"
	droneSrclibBasicImage      = "sourcegraph/srclib-basic:6c458fb05d587e521c0aad930bf7995f74a28d56"
	droneSrclibCSSImage        = "sourcegraph/srclib-css:3db83c4a7eef5d368a5e32628c93169e01d1e58a"
	droneSrclibCSharpImage     = "sourcegraph/srclib-csharp:36fe93491510033fdb483bc37171b9161ac37e78"
	droneSrclibGoImage         = "sourcegraph/srclib-go:2bf78d43ec53d8fd81c77392b8ebe2f4181a4fd6"
	droneSrclibJSONImage       = "sourcegraph/srclib-json:da879fdce0b22a422c9607f054a9f8f03b2c69f5"
	droneSrclibJavaImage       = "sourcegraph/srclib-java:2fb7ab2648527bbe4ba16812558ac9be64da8098"
	droneSrclibJavaScriptImage = "sourcegraph/srclib-javascript:3adeecd47155b75ea6d7836dc1035bc11e059876"
	droneSrclibPythonImage     = "sourcegraph/srclib-python:a65ca94fa6e09eb9a752d4d6c6c4bd5e06f83699"
	droneSrclibTypeScriptImage = "sourcegraph/srclib-typescript:04e5b07b925a646ecc296535fa6d7b9d010c53bb"
	droneSrclibCtagsImage      = "sourcegraph/srclib-ctags:13b39a491ad17b68464c69733ab95268d6541ad5"
)

func versionHash(image string) (string, error) {
	if strings.Contains(image, "@sha256:") {
		// Images built with docker 1.10 and above can not be pulled
		// with the sha256 style tags.
		return "", fmt.Errorf("we do not support '@sha256:' style tags from toolchain image %s", image)
	}
	split := strings.Split(image, ":")
	if len(split) != 2 {
		return "", fmt.Errorf("cannot parse version hash from toolchain image %s", image)
	}
	return split[1], nil
}

func SrclibVersion(lang string) (string, error) {
	switch lang {
	case "Bash":
		return versionHash(droneSrclibBashImage)
	case "Go":
		return versionHash(droneSrclibGoImage)
	case "JavaScript":
		return versionHash(droneSrclibJavaScriptImage)
	case "Java":
		return versionHash(droneSrclibJavaImage)
	case "TypeScript":
		return versionHash(droneSrclibTypeScriptImage)
	case "C#":
		return versionHash(droneSrclibCSharpImage)
	case "CSS":
		return versionHash(droneSrclibCSSImage)
	case "Python":
		return versionHash(droneSrclibPythonImage)
	case "JSON":
		return versionHash(droneSrclibJSONImage)
	case "Ctags":
		return versionHash(droneSrclibCtagsImage)
	}
	return "", fmt.Errorf("no srclib image found for %s", lang)
}
