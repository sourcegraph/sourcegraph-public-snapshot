load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")

DOCSITE_VERSION = "1.9.2"
SRC_CLI_VERSION = "5.1.0"
CTAGS_VERSION = "5.9.20220403.0"

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
        sha256 = "ccf3b2d37665864ad2d3327de71d3d0b3b3d6ef61ac30c3e052ffffd5517fa0e",
        executable = True,
    )

    http_file(
        name = "docsite_darwin_arm64",
        urls = ["https://github.com/sourcegraph/docsite/releases/download/v{0}/docsite_v{0}_darwin_arm64".format(DOCSITE_VERSION)],
        sha256 = "a44b755a0e06e18c18cd81874c495ca9b358687f77886476f641b61c1ce4937a",
        executable = True,
    )

    http_file(
        name = "docsite_linux_amd64",
        urls = ["https://github.com/sourcegraph/docsite/releases/download/v{0}/docsite_v{0}_linux_amd64".format(DOCSITE_VERSION)],
        sha256 = "ef69bd376fd1bb13dc2f72cd499c0c70510de8b4778d3210d6899f88955fc982",
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
        sha256 = "aasdfasdf",
        url = "https://github.com/sourcegraph/src-cli/releases/download/{0}/src-cli_{0}_darwin_amd64.tar.gz".format(SRC_CLI_VERSION),
    )

    http_archive(
        name = "src-cli-darwin-arm64",
        build_file_content = SRC_CLI_BUILDFILE.format("darwin-arm64"),
        sha256 = "93dc6c8522792ea16e3c8c81c8cf655a908118e867fda43c048c9b51f4c70e88",
        url = "https://github.com/sourcegraph/src-cli/releases/download/{0}/src-cli_{0}_darwin_arm64.tar.gz".format(SRC_CLI_VERSION),
    )

    # universal-ctags #
    http_file(
        name = "universal-ctags-darwin-x86_64",
        url = "https://storage.googleapis.com/universal_ctags/x86_64-darwin/bin/universal-ctags-{0}".format(CTAGS_VERSION),
        executable = True,
    )

    http_file(
        name = "universal-ctags-darwin-arm64",
        sha256 = "51b3b7ea296455e00fc5a7aafea49bb89551e81770b3728c97a04a5614fde8c5",
        url = "https://storage.googleapis.com/universal_ctags/aarch64-darwin/bin/universal-ctags-{0}".format(CTAGS_VERSION),
        executable = True,
    )

    http_file(
        name = "universal-ctags-linux-amd64",
        sha256 = "1d349d15736a30c9cc18c1fd9efbfc6081fb59125d799b84cef6b34c735fa28a",
        url = "https://storage.googleapis.com/universal_ctags/x86_64-linux/bin/universal-ctags-{0}".format(CTAGS_VERSION),
        executable = True,
    )
