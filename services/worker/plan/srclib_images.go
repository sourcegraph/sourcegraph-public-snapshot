package plan

import (
	"fmt"
	"strings"
)

// Drone Docker images for running each toolchain. Remove the sha256 version when
// developing to make it easier to test out changes to a given toolchain. E.g.,
// `droneSrclibGoImage = "sourcegraph/srclib-go"`.
var (
	droneSrclibBashImage       = "sourcegraph/srclib-bash@sha256:641dddeee4ec7db91ed59af78f2f936bdce451ea7b496b0319f20ebe6dfba255"
	droneSrclibGoImage         = "sourcegraph/srclib-go@sha256:4c91ca8b3d7fc123e9489e6751c94791776d303e21eb11e2cb14d986b78b1b06"
	droneSrclibJavaScriptImage = "sourcegraph/srclib-javascript@sha256:7cfc4ea50aaf0fea46b8704e80ef50dfa45afe82af57e653bae34ae56288a859"
	droneSrclibJavaImage       = "sourcegraph/srclib-java@sha256:4f3ffd9dd8b4b0f38c9c2b381dd5dc2103b5052e0d1524f7a0c7883b122e5056"
	droneSrclibTypeScriptImage = "sourcegraph/srclib-typescript@sha256:39adfea4bdaea50be63431fe8c85c174a6a83d34db1196ac0bb171cb79cc88d6"
	droneSrclibCSharpImage     = "sourcegraph/srclib-csharp@sha256:e5c112fc5ccb0551a09289cd732b00d038c4bc366f853e552826e36bcd903507"
	droneSrclibCSSImage        = "sourcegraph/srclib-css@sha256:5dea2ffe7183d2fb3f3f3d6a43790189f1c9a82ef46d8d605d87a99830ca9fbd"
	droneSrclibPythonImage     = "sourcegraph/srclib-python@sha256:ce9625fc4648cc2200237dc95c9b14b0eeb01621a55a53f5123509e84da069be"
	droneSrclibJSONImage       = "sourcegraph/srclib-json@sha256:8c57b51ad1f0047540106d63fac2d924d0278ae421a470a3067c390ae6edb1fc"
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
	}
	return "", fmt.Errorf("no srclib image found for %s", lang)
}
