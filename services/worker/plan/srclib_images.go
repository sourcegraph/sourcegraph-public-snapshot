package plan

// Drone Docker images for running each toolchain. Remove the sha256 version when
// developing to make it easier to test out changes to a given toolchain. E.g.,
// `droneSrclibGoImage = "sourcegraph/srclib-go"`.
var (
	droneSrclibGoImage         = "sourcegraph/srclib-go:latest"
	droneSrclibJavaScriptImage = "sourcegraph/srclib-javascript:latest"
	droneSrclibJavaImage       = "sourcegraph/srclib-java:latest"
	droneSrclibBasicImage      = "sourcegraph/srclib-basic:latest"
	droneSrclibPythonImage     = "sourcegraph/srclib-python:latest"
	droneSrclibTypeScriptImage = "sourcegraph/srclib-typescript:latest"
	droneSrclibCSharpImage     = "sourcegraph/srclib-csharp:latest"
)
