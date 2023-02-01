load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "bazel_skylib",
    sha256 = "74d544d96f4a5bb630d465ca8bbcfe231e3594e5aae57e1edbf17a6eb3ca2506",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.3.0/bazel-skylib-1.3.0.tar.gz",
        "https://github.com/bazelbuild/bazel-skylib/releases/download/1.3.0/bazel-skylib-1.3.0.tar.gz",
    ],
)

load("@bazel_skylib//:workspace.bzl", "bazel_skylib_workspace")

bazel_skylib_workspace()

http_archive(
    name = "aspect_rules_js",
    sha256 = "ad666b12324bab8bc151772bb2eff9aadace7bfd4d624157c2ac3931860d1c95",
    strip_prefix = "rules_js-1.11.1",
    url = "https://github.com/aspect-build/rules_js/archive/refs/tags/v1.11.1.tar.gz",
)

http_archive(
    name = "rules_nodejs",
    sha256 = "08337d4fffc78f7fe648a93be12ea2fc4e8eb9795a4e6aa48595b66b34555626",
    urls = ["https://github.com/bazelbuild/rules_nodejs/releases/download/5.8.0/rules_nodejs-core-5.8.0.tar.gz"],
)

http_archive(
    name = "aspect_rules_ts",
    sha256 = "e81f37c4fe014fc83229e619360d51bfd6cb8ac405a7e8018b4a362efa79d000",
    strip_prefix = "rules_ts-1.0.4",
    url = "https://github.com/aspect-build/rules_ts/archive/refs/tags/v1.0.4.tar.gz",
)

http_archive(
    name = "aspect_rules_jest",
    sha256 = "f2be891d9b38473f08a336e5f6ca327bfdc90411b1798a1c476c6f6ceae54520",
    strip_prefix = "rules_jest-0.13.1",
    url = "https://github.com/aspect-build/rules_jest/archive/refs/tags/v0.13.1.tar.gz",
)

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "56d8c5a5c91e1af73eca71a6fab2ced959b67c86d12ba37feedb0a2dfea441a6",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.37.0/rules_go-v0.37.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.37.0/rules_go-v0.37.0.zip",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "448e37e0dbf61d6fa8f00aaa12d191745e14f07c31cabfa731f0c8e8a4f41b97",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.28.0/bazel-gazelle-v0.28.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.28.0/bazel-gazelle-v0.28.0.tar.gz",
    ],
)

http_archive(
    name = "com_google_protobuf",
    sha256 = "6aff9834fd7c540875e1836967c8d14c6897e3785a2efac629f69860fb7834ff",
    strip_prefix = "protobuf-3.15.0",
    urls = [
        "https://github.com/protocolbuffers/protobuf/archive/v3.15.0.tar.gz",
    ],
)

# Node toolchain setup ==========================
load("@rules_nodejs//nodejs:repositories.bzl", "DEFAULT_NODE_VERSION", "nodejs_register_toolchains")

nodejs_register_toolchains(
    name = "nodejs",
    node_version = DEFAULT_NODE_VERSION,
)

# rules_js setup ================================
load("@aspect_rules_js//js:repositories.bzl", "rules_js_dependencies")

rules_js_dependencies()

# rules_js npm setup ============================
load("@aspect_rules_js//npm:npm_import.bzl", "npm_translate_lock")

npm_translate_lock(
    name = "npm",
    npmrc = "//:.npmrc",
    pnpm_lock = "//:pnpm-lock.yaml",
    verify_node_modules_ignored = "//:.bazelignore",
)

# rules_ts npm setup ============================
load("@npm//:repositories.bzl", "npm_repositories")

npm_repositories()

load("@aspect_rules_ts//ts:repositories.bzl", "rules_ts_dependencies", LATEST_TS_VERSION = "LATEST_VERSION")

rules_ts_dependencies(ts_version = LATEST_TS_VERSION)

# rules_jest setup ==============================
load("@aspect_rules_jest//jest:dependencies.bzl", "rules_jest_dependencies")

rules_jest_dependencies()

load("@aspect_rules_jest//jest:repositories.bzl", "jest_repositories")

jest_repositories(name = "jest")

load("@jest//:npm_repositories.bzl", jest_npm_repositories = "npm_repositories")

jest_npm_repositories()

# Go toolchain setup

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
load("//:deps.bzl", "go_dependencies")

# gazelle:repository_macro deps.bzl%go_dependencies
go_dependencies()

go_rules_dependencies()

go_register_toolchains(version = "1.19.3")

gazelle_dependencies()

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()
