workspace(name = "com_github_mwitkow_go_proto_validators")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

GO_VERSION = "1.14.7"

http_archive(
    name = "bazel_skylib",
    sha256 = "64ad2728ccdd2044216e4cec7815918b7bb3bb28c95b7e9d951f9d4eccb07625",
    strip_prefix = "bazel-skylib-1.0.2",
    url = "https://github.com/bazelbuild/bazel-skylib/archive/1.0.2.zip",
)

load("@bazel_skylib//:workspace.bzl", "bazel_skylib_workspace")

bazel_skylib_workspace()

http_archive(
    name = "rules_proto",
    sha256 = "296ffd3e7992bd83fa75151255f7c7f27d22d6e52e2fd3c3d3d10c292317fbed",
    strip_prefix = "rules_proto-f6c112fa4eb2b8f934feb938a6fce41425e41587",
    urls = [
        "https://github.com/bazelbuild/rules_proto/archive/f6c112fa4eb2b8f934feb938a6fce41425e41587.tar.gz",
    ],
)

http_archive(
    name = "com_google_protobuf",
    sha256 = "e5265d552e12c1f39c72842fa91d84941726026fa056d914ea6a25cd58d7bbf8",
    strip_prefix = "protobuf-3.12.3",
    url = "https://github.com/protocolbuffers/protobuf/archive/v3.12.3.zip",
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "0310e837aed522875791750de44408ec91046c630374990edd51827cb169f616",
    urls = [
        "https://github.com/bazelbuild/rules_go/releases/download/v0.23.7/rules_go-v0.23.7.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(go_version = GO_VERSION)

http_archive(
    name = "bazel_gazelle",
    sha256 = "2423201f91471ea87925b81962258e27a22cd8ebb4fe355bf033dcf2ad668541",
    strip_prefix = "bazel-gazelle-0.21.1",
    urls = [
        "https://github.com/bazelbuild/bazel-gazelle/archive/v0.21.1.tar.gz",
    ],
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

load("go_deps.bzl", "go_repositories")

go_repositories()

# gazelle:repository_macro go_deps.bzl%go_repositories
