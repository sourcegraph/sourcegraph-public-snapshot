load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")

DOCSITE_VERSION = "1.9.4"
SRC_CLI_VERSION = "5.1.0"
CTAGS_VERSION = "6.0.0.2783f009"

SRC_CLI_BUILDFILE = """
filegroup(
    name = "src-cli-{}",
    srcs = ["src"],
    visibility = ["//visibility:public"],
)
"""

def tool_deps():
    # Docsite #
    http_file(
        name = "docsite_darwin_amd64",
        urls = ["https://github.com/sourcegraph/docsite/releases/download/v{0}/docsite_v{0}_darwin_amd64".format(DOCSITE_VERSION)],
        sha256 = "f3ad94e1398cc30e45518c82bd6fa9f7c386a8c395811ba49def24113215a2d9",
        executable = True,
    )

    http_file(
        name = "docsite_darwin_arm64",
        urls = ["https://github.com/sourcegraph/docsite/releases/download/v{0}/docsite_v{0}_darwin_arm64".format(DOCSITE_VERSION)],
        sha256 = "b817d794537f38720d5c07eb323e729391b2d4ff85d1dac691e3cfe7a3cb6d13",
        executable = True,
    )

    http_file(
        name = "docsite_linux_amd64",
        urls = ["https://github.com/sourcegraph/docsite/releases/download/v{0}/docsite_v{0}_linux_amd64".format(DOCSITE_VERSION)],
        sha256 = "7d60a55eb5017ebeb3a523143bd3007f74297db685491fad84499b2c60b3a872",
        executable = True,
    )

    # src-cli #
    http_archive(
        name = "src-cli-linux-amd64",
        build_file_content = SRC_CLI_BUILDFILE.format("linux-amd64"),
        sha256 = "270ddad7748c1b76f082b637e336b5c7a58af76d207168469f4b7bef957953e3",
        url = "https://github.com/sourcegraph/src-cli/releases/download/{0}/src-cli_{0}_linux_amd64.tar.gz".format(SRC_CLI_VERSION),
    )

    http_archive(
        name = "src-cli-darwin-amd64",
        build_file_content = SRC_CLI_BUILDFILE.format("darwin-amd64"),
        sha256 = "f14414e3ff4759cd1fbed0107138214f87d9a69cdb55ed1c4522704069420d9b",
        url = "https://github.com/sourcegraph/src-cli/releases/download/{0}/src-cli_{0}_darwin_amd64.tar.gz".format(SRC_CLI_VERSION),
    )

    http_archive(
        name = "src-cli-darwin-arm64",
        build_file_content = SRC_CLI_BUILDFILE.format("darwin-arm64"),
        sha256 = "93dc6c8522792ea16e3c8c81c8cf655a908118e867fda43c048c9b51f4c70e88",
        url = "https://github.com/sourcegraph/src-cli/releases/download/{0}/src-cli_{0}_darwin_arm64.tar.gz".format(SRC_CLI_VERSION),
    )

    # universal-ctags
    #
    # Two step process to update these. First land a commit in main updating
    # the version in dev/nix/ctags.nix. Then copy the hashes from
    # https://github.com/sourcegraph/sourcegraph/actions/workflows/universal-ctags.yml
    http_file(
        name = "universal-ctags-darwin-amd64",
        sha256 = "7aead221c07d8092a0cbfad6e69ae526292e405bbe2f06d38d346969e23a6f68",
        url = "https://storage.googleapis.com/universal_ctags/x86_64-darwin/dist/universal-ctags-{0}".format(CTAGS_VERSION),
        executable = True,
    )

    http_file(
        name = "universal-ctags-darwin-arm64",
        sha256 = "ac4a73b69042c60e68f80f8819d5c6b05e233386042ba867205a252046d6471e",
        url = "https://storage.googleapis.com/universal_ctags/aarch64-darwin/dist/universal-ctags-{0}".format(CTAGS_VERSION),
        executable = True,
    )

    http_file(
        name = "universal-ctags-linux-amd64",
        sha256 = "e99b942754ce9d55c9445513236c010753d26decbb38ed932780bec098ac0809",
        url = "https://storage.googleapis.com/universal_ctags/x86_64-linux/dist/universal-ctags-{0}".format(CTAGS_VERSION),
        executable = True,
    )
