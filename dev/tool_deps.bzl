load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

DOCSITE_VERSION="1.9.2"

def tool_deps():
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

