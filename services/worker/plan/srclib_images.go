package plan

// Drone Docker images for running each toolchain. Remove the sha256 version when
// developing to make it easier to test out changes to a given toolchain. E.g.,
// `droneSrclibGoImage = "sourcegraph/srclib-go"`.
var (
	droneSrclibGoImage         = "sourcegraph/srclib-go@sha256:5b54a44e5790f3adfe4b7a31db167d52e35513adc833c4d4ca2b317cdee30f08"
	droneSrclibJavaScriptImage = "sourcegraph/srclib-javascript@sha256:09621a45720701482a6cc5c113a58e167a9c5b8e265ba82b07af44c84a61846f"
	droneSrclibJavaImage       = "sourcegraph/srclib-java@sha256:8b1fdad37e8daae89582dc7c079022e15b332e0ee93f1b35e60d5b16ee2dd38a"
	droneSrclibBasicImage      = "sourcegraph/srclib-basic@sha256:4157bcbec38ed83dde449ebad68f753ff55908956f684e0f1645bf4afa785792"
	droneSrclibPythonImage     = "sourcegraph/srclib-python@sha256:0a5cb930bd9aa320f98a2b920b552b3e4027daf04e4bfaf1f4a9c2a61d68d561"
	droneSrclibTypeScriptImage = "sourcegraph/srclib-typescript@sha256:4c7a507d9c1d25bd8379613d2dfd4e41dd63d2098a9804d1b57fa05854f9414c"
	droneSrclibCSharpImage     = "sourcegraph/srclib-csharp@sha256:f899fcc4b73f6005c0cf3b9050e7b83299c2b2616f8401d8d00b6c1ced5fecd9"
)
