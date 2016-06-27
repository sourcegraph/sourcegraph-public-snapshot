package plan

import (
	"fmt"
	"strings"
)

// Drone Docker images for running each toolchain. Remove the sha256 version when
// developing to make it easier to test out changes to a given toolchain. E.g.,
// `droneSrclibGoImage = "sourcegraph/srclib-go"`.
var (
	droneSrclibBashImage       = "sourcegraph/srclib-bash@sha256:2b5468bd134f34c32ef17af27ecb10561fd1d064104dfad8bf3a0263ffad3dab"
	droneSrclibGoImage         = "sourcegraph/srclib-go:2bf78d43ec53d8fd81c77392b8ebe2f4181a4fd6"
	droneSrclibJavaScriptImage = "sourcegraph/srclib-javascript@sha256:fd680ae719f2ef4ca82061ee3d0dc062ce1577711e8178a728e1be2ac1526193"
	droneSrclibJavaImage       = "sourcegraph/srclib-java:2fb7ab2648527bbe4ba16812558ac9be64da8098"
	droneSrclibTypeScriptImage = "sourcegraph/srclib-typescript@sha256:39adfea4bdaea50be63431fe8c85c174a6a83d34db1196ac0bb171cb79cc88d6"
	droneSrclibCSharpImage     = "sourcegraph/srclib-csharp@sha256:faa3c210e22693dc33954fb9d6714c7a735372e3a90a4cebc19c64be42551177"
	droneSrclibCSSImage        = "sourcegraph/srclib-css@sha256:fb756d7443c72350c65f2141675efddbeb611603b77ee11f0ebaf03150e979a5"
	droneSrclibPythonImage     = "sourcegraph/srclib-python@sha256:d9b612e504c06ac882e5b4e34339d536a8119f8527330573ea10b1b625324fa2"
	droneSrclibJSONImage       = "sourcegraph/srclib-json@sha256:8c57b51ad1f0047540106d63fac2d924d0278ae421a470a3067c390ae6edb1fc"
)

func versionHash(image string) (string, error) {
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
	}
	return "", fmt.Errorf("no srclib image found for %s", lang)
}
