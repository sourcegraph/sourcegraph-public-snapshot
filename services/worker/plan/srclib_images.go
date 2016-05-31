package plan

import (
	"fmt"
	"strings"
)

// Drone Docker images for running each toolchain. Remove the sha256 version when
// developing to make it easier to test out changes to a given toolchain. E.g.,
// `droneSrclibGoImage = "sourcegraph/srclib-go"`.
var (
	droneSrclibGoImage         = "sourcegraph/srclib-go@sha256:144e8bf8bb3d8128b2307594f624ca8eaf8f552ce34f62c0480cdc5100ba6e90"
	droneSrclibJavaScriptImage = "sourcegraph/srclib-javascript@sha256:d09a1d9cb01f27fefccb13532df09e581e3ed23d924cc50e8d281dd4ad47e275"
	droneSrclibJavaImage       = "sourcegraph/srclib-java@sha256:debb06b813143b02d35be8070c3ff89e0c68026332e14f9ededd7afd592c0b6c"
	droneSrclibTypeScriptImage = "sourcegraph/srclib-typescript@sha256:39adfea4bdaea50be63431fe8c85c174a6a83d34db1196ac0bb171cb79cc88d6"
	droneSrclibCSharpImage     = "sourcegraph/srclib-csharp@sha256:e5c112fc5ccb0551a09289cd732b00d038c4bc366f853e552826e36bcd903507"
	droneSrclibCSSImage        = "sourcegraph/srclib-css@sha256:c8b71da5d2211adb4a4c44ceef1ca4e19d2db42fa769a7aeeaed56d3cd6040ff"
	droneSrclibPythonImage     = "sourcegraph/srclib-python@sha256:aeb6ecf215d75936a539ca72770ef0dd996cd2ebd0e6130a4deb32a4edf6a48a"
)

func versionHash(image string) (string, error) {
	split := strings.Split(image, "@sha256:")
	if len(split) != 2 {
		return "", fmt.Errorf("cannot parse version hash from toolchain image %s", image)
	}

	return split[1], nil
}

func SrclibVersion(lang string) (string, error) {
	switch lang {
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
	}

	return "", fmt.Errorf("no srclib image found for %s", lang)
}
