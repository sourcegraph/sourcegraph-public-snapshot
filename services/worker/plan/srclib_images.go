package plan

// Drone Docker images for running each toolchain. Remove the sha256 version when
// developing to make it easier to test out changes to a given toolchain. E.g.,
// `droneSrclibGoImage = "sourcegraph/srclib-go"`.
var (
	droneSrclibGoImage         = "sourcegraph/srclib-go@sha256:f9ef9fac78399b4da7a7209bd2f76c0532ebfc67585039268732f51f7ee09d39"
	droneSrclibJavaScriptImage = "sourcegraph/srclib-javascript@sha256:6dd3fbad7b5cb4ae897d2b8ff88747321642028b15c77c152b8828971ec72601"
	droneSrclibJavaImage       = "sourcegraph/srclib-java@sha256:a16ae88df7fbc54f96a0e38d32686b05af5cc55f71d018b1640bb3747b8e11df"
	droneSrclibTypeScriptImage = "sourcegraph/srclib-typescript@sha256:4c7a507d9c1d25bd8379613d2dfd4e41dd63d2098a9804d1b57fa05854f9414c"
	droneSrclibCSharpImage     = "sourcegraph/srclib-csharp@sha256:20850a4beeaf56e3b398b714294371b5e77f8139b903b0e07218e2964dad9afa"
	droneSrclibCSSImage        = "sourcegraph/srclib-css@sha256:7d619b5ceac0198b7f1911f2f535eda3e037b1489c52293090e6000093346987"
)
