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
    name = "aspect_bazel_lib",
    sha256 = "2518c757715d4f5fc7cc7e0a68742dd1155eaafc78fb9196b8a18e13a738cea2",
    strip_prefix = "bazel-lib-1.28.0",
    url = "https://github.com/aspect-build/bazel-lib/releases/download/v1.28.0/bazel-lib-v1.28.0.tar.gz",
)

http_archive(
    name = "aspect_rules_js",
    sha256 = "a592fafd8a27b2828318cebbda0003686c6da3318df366b563e8beeffa05a02c",
    strip_prefix = "rules_js-1.21.0",
    url = "https://github.com/aspect-build/rules_js/releases/download/v1.21.0/rules_js-v1.21.0.tar.gz",
)

http_archive(
    name = "aspect_rules_ts",
    sha256 = "02480b6a1cd12516edf364e678412e9da10445fe3f1070c014ac75e922c969ea",
    strip_prefix = "rules_ts-1.3.1",
    url = "https://github.com/aspect-build/rules_ts/releases/download/v1.3.1/rules_ts-v1.3.1.tar.gz",
)

http_archive(
    name = "aspect_rules_jest",
    sha256 = "bf8f4a4d2a833e4f96f866c686c38bcee69d3bdae8a827b1c9d2fdf92212bc0b",
    strip_prefix = "rules_jest-95d8f1961a9c6f3aee2929881b1b74461652e775",
    url = "https://github.com/aspect-build/rules_jest/archive/95d8f1961a9c6f3aee2929881b1b74461652e775.tar.gz",
)

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "dd926a88a564a9246713a9c00b35315f54cbd46b31a26d5d8fb264c07045f05d",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.38.1/rules_go-v0.38.1.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.38.1/rules_go-v0.38.1.zip",
    ],
)

http_archive(
    name = "rules_buf",
    sha256 = "523a4e06f0746661e092d083757263a249fedca535bd6dd819a8c50de074731a",
    strip_prefix = "rules_buf-0.1.1",
    urls = [
        "https://github.com/bufbuild/rules_buf/archive/refs/tags/v0.1.1.zip",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "ecba0f04f96b4960a5b250c8e8eeec42281035970aa8852dda73098274d14a1d",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.29.0/bazel-gazelle-v0.29.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.29.0/bazel-gazelle-v0.29.0.tar.gz",
    ],
)

http_archive(
    name = "rules_rust",
    sha256 = "dc8d79fe9a5beb79d93e482eb807266a0e066e97a7b8c48d43ecf91f32a3a8f3",
    urls = ["https://github.com/bazelbuild/rules_rust/releases/download/0.19.0/rules_rust-v0.19.0.tar.gz"],
)

http_archive(
    name = "io_tweag_rules_nixpkgs",
    sha256 = "cb1030a6134f625e2d30d2a34dcfe7960157ae21ec8f20c2b1adb0665f789f50",
    strip_prefix = "rules_nixpkgs-4dddbafba508cd2dffd95b8562cab91c9336fe36",
    urls = ["https://github.com/tweag/rules_nixpkgs/archive/4dddbafba508cd2dffd95b8562cab91c9336fe36.tar.gz"],
)

# rules_js setup ================================
load("@aspect_rules_js//js:repositories.bzl", "rules_js_dependencies")

rules_js_dependencies()

# node toolchain setup ==========================
load("@rules_nodejs//nodejs:repositories.bzl", "nodejs_register_toolchains")

nodejs_register_toolchains(
    name = "nodejs",
    node_version = "16.19.0",
)

# rules_js npm setup ============================
load("@aspect_rules_js//npm:npm_import.bzl", "npm_translate_lock")

npm_translate_lock(
    name = "npm",
    npm_package_target_name = "{dirname}_pkg",
    npmrc = "//:.npmrc",
    pnpm_lock = "//:pnpm-lock.yaml",
    verify_node_modules_ignored = "//:.bazelignore",
)

# rules_ts npm setup ============================
load("@npm//:repositories.bzl", "npm_repositories")

npm_repositories()

load("@aspect_rules_ts//ts:repositories.bzl", "rules_ts_dependencies")

rules_ts_dependencies(ts_version = "4.9.5")

# rules_jest setup ==============================
load("@aspect_rules_jest//jest:dependencies.bzl", "rules_jest_dependencies")

rules_jest_dependencies()

load("@aspect_rules_jest//jest:repositories.bzl", "jest_repositories")

jest_repositories(
    name = "jest",
    jest_version = "v28.1.0",
)

load("@jest//:npm_repositories.bzl", jest_npm_repositories = "npm_repositories")

jest_npm_repositories()

# rules_esbuild setup ===========================
http_archive(
    name = "aspect_rules_esbuild",
    sha256 = "621c8ccb8a1400951c52357377eda4c575c6c901689bb629969881e0be8a8614",
    strip_prefix = "rules_esbuild-175023fede7d2532dd35d89cb43b43cbe9e75678",
    url = "https://github.com/aspect-build/rules_esbuild/archive/175023fede7d2532dd35d89cb43b43cbe9e75678.tar.gz",
)

load("@aspect_rules_esbuild//esbuild:dependencies.bzl", "rules_esbuild_dependencies")

rules_esbuild_dependencies()

# Register a toolchain containing esbuild npm package and native bindings
load("@aspect_rules_esbuild//esbuild:repositories.bzl", "LATEST_VERSION", "esbuild_register_toolchains")

esbuild_register_toolchains(
    name = "esbuild",
    esbuild_version = LATEST_VERSION,
)

# rules_webpack setup ===========================
# Commit to include unreleased https://github.com/aspect-build/rules_webpack/commit/4a5f04a4bc504f71d32825124c7872ff721aa1b0
http_archive(
    name = "aspect_rules_webpack",
    sha256 = "8d81f8d018127c72270ea4b7287be5c4ff63d9656a34334c305d52f14e0c922f",
    strip_prefix = "rules_webpack-4a5f04a4bc504f71d32825124c7872ff721aa1b0",
    url = "https://github.com/aspect-build/rules_webpack/archive/4a5f04a4bc504f71d32825124c7872ff721aa1b0.tar.gz",
)

load("@aspect_rules_webpack//webpack:dependencies.bzl", "rules_webpack_dependencies")

rules_webpack_dependencies()

load("@aspect_rules_webpack//webpack:repositories.bzl", "webpack_repositories")

webpack_repositories(name = "webpack")

load("@webpack//:npm_repositories.bzl", webpack_npm_repositories = "npm_repositories")

webpack_npm_repositories()

# Go toolchain setup

load("@rules_buf//buf:repositories.bzl", "rules_buf_dependencies", "rules_buf_toolchains")

rules_buf_dependencies()

rules_buf_toolchains(version = "v1.11.0")

load("@rules_buf//gazelle/buf:repositories.bzl", "gazelle_buf_dependencies")

gazelle_buf_dependencies()

load("@rules_proto//proto:repositories.bzl", "rules_proto_dependencies", "rules_proto_toolchains")

rules_proto_dependencies()

rules_proto_toolchains()

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")
load("//:linter_deps.bzl", "linter_dependencies")
load("//:deps.bzl", "go_dependencies")

go_repository(
    name = "com_github_aws_aws_sdk_go_v2_service_ssooidc",
    build_file_proto_mode = "disable_global",
    importpath = "github.com/aws/aws-sdk-go-v2/service/ssooidc",
    sum = "h1:0bLhH6DRAqox+g0LatcjGKjjhU6Eudyys6HB6DJVPj8=",
    version = "v1.14.1",
)

# gazelle:repository_macro deps.bzl%go_dependencies
go_dependencies()

go_rules_dependencies()

go_register_toolchains(
    nogo = "@//:sg_nogo",
    version = "1.19.8",
)

linter_dependencies()

gazelle_dependencies()

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

# rust toolchain setup
load("@rules_rust//rust:repositories.bzl", "rules_rust_dependencies", "rust_register_toolchains")

rules_rust_dependencies()

rust_register_toolchains(
    edition = "2021",
    # Keep in sync with docker-images/syntax-highlighter/Dockerfile
    # and docker-images/syntax-highlighter/rust-toolchain.toml
    versions = [
        "1.68.0",
    ],
)

load("@rules_rust//crate_universe:defs.bzl", "crates_repository")

crates_repository(
    name = "crate_index",
    cargo_config = "//docker-images/syntax-highlighter:.cargo/config.toml",
    cargo_lockfile = "//docker-images/syntax-highlighter:Cargo.lock",
    # this file has to be manually created and it will be filled when
    # the target is ran.
    # To regenerate this file run: CARGO_BAZEL_REPIN=1 bazel sync --only=crate_index
    lockfile = "//docker-images/syntax-highlighter:Cargo.Bazel.lock",
    # glob doesn't work in WORKSPACE files: https://github.com/bazelbuild/bazel/issues/11935
    manifests = [
        "//docker-images/syntax-highlighter:Cargo.toml",
        "//docker-images/syntax-highlighter:crates/scip-macros/Cargo.toml",
        "//docker-images/syntax-highlighter:crates/scip-syntax/Cargo.toml",
        "//docker-images/syntax-highlighter:crates/scip-treesitter/Cargo.toml",
        "//docker-images/syntax-highlighter:crates/scip-treesitter-languages/Cargo.toml",
        "//docker-images/syntax-highlighter:crates/sg-syntax/Cargo.toml",
    ],
)

load("@crate_index//:defs.bzl", "crate_repositories")

crate_repositories()

BAZEL_ZIG_CC_VERSION = "v1.0.1"

http_archive(
    name = "bazel-zig-cc",
    sha256 = "e9f82bfb74b3df5ca0e67f4d4989e7f1f7ce3386c295fd7fda881ab91f83e509",
    strip_prefix = "bazel-zig-cc-{}".format(BAZEL_ZIG_CC_VERSION),
    urls = [
        "https://mirror.bazel.build/github.com/uber/bazel-zig-cc/releases/download/{0}/{0}.tar.gz".format(BAZEL_ZIG_CC_VERSION),
        "https://github.com/uber/bazel-zig-cc/releases/download/{0}/{0}.tar.gz".format(BAZEL_ZIG_CC_VERSION),
    ],
)

load("@bazel-zig-cc//toolchain:defs.bzl", zig_toolchains = "toolchains")

zig_toolchains()

load("//dev/backcompat:defs.bzl", "back_compat_defs")

back_compat_defs()

# nixos toolchains setup ===============================
load("@io_tweag_rules_nixpkgs//nixpkgs:repositories.bzl", "rules_nixpkgs_dependencies")

rules_nixpkgs_dependencies(toolchains = [
    "nodejs",
    "rust",
])

load("//dev:nix.bzl", "nix_deps")

nix_deps()
