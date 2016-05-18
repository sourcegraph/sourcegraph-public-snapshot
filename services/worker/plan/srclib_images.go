package plan

// Drone Docker images for running each toolchain. Remove the sha256 version when
// developing to make it easier to test out changes to a given toolchain. E.g.,
// `droneSrclibGoImage = "sourcegraph/srclib-go"`.
var (
	droneSrclibGoImage         = "sourcegraph/srclib-go@sha256:8fcf921d22dfd87215d0cd545dd65d21a26b46ce8b4cba3c66793d975312ff28"
	droneSrclibJavaScriptImage = "sourcegraph/srclib-javascript@sha256:6dd3fbad7b5cb4ae897d2b8ff88747321642028b15c77c152b8828971ec72601"
	droneSrclibJavaImage       = "sourcegraph/srclib-java@sha256:1aa7f5ed4417c928eb0226a4d2f4eedb11bea8222a54b1838b92265979fd94fa"
	droneSrclibTypeScriptImage = "sourcegraph/srclib-typescript@sha256:4c7a507d9c1d25bd8379613d2dfd4e41dd63d2098a9804d1b57fa05854f9414c"
	droneSrclibCSharpImage     = "sourcegraph/srclib-csharp@sha256:20850a4beeaf56e3b398b714294371b5e77f8139b903b0e07218e2964dad9afa"
	droneSrclibCSSImage        = "sourcegraph/srclib-css@sha256:7d619b5ceac0198b7f1911f2f535eda3e037b1489c52293090e6000093346987"
)
