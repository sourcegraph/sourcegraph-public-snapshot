package plan

// Drone Docker images for running each toolchain. Remove the sha256 version when
// developing to make it easier to test out changes to a given toolchain. E.g.,
// `droneSrclibGoImage = "sourcegraph/srclib-go"`.
var (
	droneSrclibGoImage         = "sourcegraph/srclib-go@sha256:d005593603afc3de91e70e331d2fee4aba6486dadb36c433799d820bdc672090"
	droneSrclibJavaScriptImage = "sourcegraph/srclib-javascript@sha256:09621a45720701482a6cc5c113a58e167a9c5b8e265ba82b07af44c84a61846f"
	droneSrclibJavaImage       = "sourcegraph/srclib-java@sha256:8b1fdad37e8daae89582dc7c079022e15b332e0ee93f1b35e60d5b16ee2dd38a"
	droneSrclibBasicImage      = "sourcegraph/srclib-basic@sha256:4157bcbec38ed83dde449ebad68f753ff55908956f684e0f1645bf4afa785792"
	droneSrclibPythonImage     = "sourcegraph/srclib-python@sha256:0a5cb930bd9aa320f98a2b920b552b3e4027daf04e4bfaf1f4a9c2a61d68d561"
	droneSrclibTypeScriptImage = "sourcegraph/srclib-typescript@sha256:4c7a507d9c1d25bd8379613d2dfd4e41dd63d2098a9804d1b57fa05854f9414c"
	droneSrclibCSharpImage     = "sourcegraph/srclib-csharp@sha256:d29301bafeab928a12bbdffa5c547ee3befccfbe7402121d7d9339ceaf57d589"
)
