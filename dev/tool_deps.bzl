"Third party dev tooling dependencies"

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")

DOCSITE_VERSION = "1.9.4"
SRC_CLI_VERSION = "5.2.1"
CTAGS_VERSION = "6.0.0.2783f009"
P4_FUSION_VERSION = "v1.13.2-sg.04a293a"

SRC_CLI_BUILDFILE = """
filegroup(
    name = "src-cli-{}",
    srcs = ["src"],
    visibility = ["//visibility:public"],
)
"""

GCLOUD_VERSION = "415.0.0"
GCLOUD_BUILDFILE = """package(default_visibility = ["//visibility:public"])\nexports_files(["gcloud", "gsutil", "bq", "git-credential-gcloud"])"""
GCLOUD_PATCH_CMDS = [
    "ln -s google-cloud-sdk/bin/gcloud gcloud",
    "ln -s google-cloud-sdk/bin/gsutil gsutil",
    "ln -s google-cloud-sdk/bin/bq bq",
    "ln -s google-cloud-sdk/bin/git-credential-gcloud.sh git-credential-gcloud",
]

def tool_deps():
    "Repository rules to fetch third party tooling used for dev purposes"

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
        sha256 = "19671ea6ee8a518fedaa45e6f6fb44767e7057c1c37dad34e36d829d5001a2f6",
        url = "https://github.com/sourcegraph/src-cli/releases/download/{0}/src-cli_{0}_linux_amd64.tar.gz".format(SRC_CLI_VERSION),
    )

    http_archive(
        name = "src-cli-darwin-amd64",
        build_file_content = SRC_CLI_BUILDFILE.format("darwin-amd64"),
        sha256 = "a05d95a05c4266e766a7ebb85078dc16c8dd1971bddf7d966cb334638ed55375",
        url = "https://github.com/sourcegraph/src-cli/releases/download/{0}/src-cli_{0}_darwin_amd64.tar.gz".format(SRC_CLI_VERSION),
    )

    http_archive(
        name = "src-cli-darwin-arm64",
        build_file_content = SRC_CLI_BUILDFILE.format("darwin-arm64"),
        sha256 = "af34afa269d29cb24b40c17bb2045e353ac6fa1c1aa1164187c8582b1538fee4",
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

    http_archive(
        name = "gcloud-darwin-arm64",
        build_file_content = GCLOUD_BUILDFILE,
        patch_cmds = GCLOUD_PATCH_CMDS,
        sha256 = "974ed4f37f8bde2f7a9731eba90b033f7c97d24d835ecc62b58eee87c8f29776",
        url = "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-{}-darwin-arm.tar.gz".format(GCLOUD_VERSION),
    )

    http_archive(
        name = "gcloud-darwin-amd64",
        build_file_content = GCLOUD_BUILDFILE,
        patch_cmds = GCLOUD_PATCH_CMDS,
        sha256 = "f05cc45ffc6c1f3ff73854989f3ea3d6bee40287d23047917e4c845aeb027f98",
        url = "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-{}-darwin-x86_64.tar.gz".format(GCLOUD_VERSION),
    )

    http_archive(
        name = "gcloud-linux-amd64",
        build_file_content = GCLOUD_BUILDFILE,
        patch_cmds = GCLOUD_PATCH_CMDS,
        sha256 = "5f9ed1862a82f393be3b16634309e9e8edb6da13a8704952be9c4c59963f9cd4",
        url = "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-{}-linux-x86_64.tar.gz".format(GCLOUD_VERSION),
    )

    http_file(
        name = "p4-fusion-linux-amd64",
        sha256 = "4c32aa00fa220733faea27a1c6ec4acd0998c1a7f870e08de9947685621f0d06",
        url = "https://storage.googleapis.com/p4-fusion/x86_64-linux/dist/p4-fusion-{0}".format(P4_FUSION_VERSION),
        executable = True,
    )

    http_file(
        name = "p4-fusion-darwin-amd64",
        sha256 = "bfa525a8a38d2c2ea205865b1a6d5be0b680e3160a64ba9505953be3294d1b9c",
        url = "https://storage.googleapis.com/p4-fusion/x86_64-darwin/dist/p4-fusion-{0}".format(P4_FUSION_VERSION),
        executable = True,
    )

    http_file(
        name = "p4-fusion-darwin-arm64",
        sha256 = "f97942e145902e682a5c1bc2608071a24d17bf943f35faaf18f359cbbaacddcd",
        url = "https://storage.googleapis.com/p4-fusion/aarch64-darwin/dist/p4-fusion-{0}".format(P4_FUSION_VERSION),
        executable = True,
    )
