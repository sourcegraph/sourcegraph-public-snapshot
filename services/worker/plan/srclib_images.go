package plan

import (
	"fmt"
	"strings"
)

// Drone Docker images for running each toolchain. Remove the sha256 version when
// developing to make it easier to test out changes to a given toolchain. E.g.,
// `droneSrclibGoImage = "sourcegraph/srclib-go"`.
var (
	droneSrclibGoImage         = "sourcegraph/srclib-go@sha256:71e923e21886dae1b0e83c2bd51aef7ddde4a6a8111b1980e8017a47f5a56a60"
	droneSrclibJavaScriptImage = "sourcegraph/srclib-javascript@sha256:6dd3fbad7b5cb4ae897d2b8ff88747321642028b15c77c152b8828971ec72601"
	droneSrclibJavaImage       = "sourcegraph/srclib-java@sha256:a9c411cd3b914504ba4d0b8c4ef023ee0794e8695089897e3e59b59ac54fc895"
	droneSrclibTypeScriptImage = "sourcegraph/srclib-typescript@sha256:a511770c9156236de59f5bb858039cd724527ba486b4e91297cd0a61fee20b4c"
	droneSrclibCSharpImage     = "sourcegraph/srclib-csharp@sha256:20850a4beeaf56e3b398b714294371b5e77f8139b903b0e07218e2964dad9afa"
	droneSrclibCSSImage        = "sourcegraph/srclib-css@sha256:7d619b5ceac0198b7f1911f2f535eda3e037b1489c52293090e6000093346987"
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
