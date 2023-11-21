load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "bazel_skylib",
    sha256 = "66ffd9315665bfaafc96b52278f57c7e2dd09f5ede279ea6d39b2be471e7e3aa",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.4.2/bazel-skylib-1.4.2.tar.gz",
        "https://github.com/bazelbuild/bazel-skylib/releases/download/1.4.2/bazel-skylib-1.4.2.tar.gz",
    ],
)

load("@bazel_skylib//:workspace.bzl", "bazel_skylib_workspace")

bazel_skylib_workspace()

http_archive(
    name = "aspect_bazel_lib",
    sha256 = "cbf473d630ab67b36461d83b38fdc44e56f45b78d03c405e4958280211124d79",
    strip_prefix = "bazel-lib-1.36.0",
    url = "https://github.com/aspect-build/bazel-lib/releases/download/v1.36.0/bazel-lib-v1.36.0.tar.gz",
)

# rules_js defines an older rules_nodejs, so we override it here
http_archive(
    name = "rules_nodejs",
    sha256 = "162f4adfd719ba42b8a6f16030a20f434dc110c65dc608660ef7b3411c9873f9",
    strip_prefix = "rules_nodejs-6.0.2",
    url = "https://github.com/bazelbuild/rules_nodejs/releases/download/v6.0.2/rules_nodejs-v6.0.2.tar.gz",
)

http_archive(
    name = "aspect_rules_js",
    sha256 = "7ab9776bcca823af361577a1a2ebb9a30d2eb5b94ecc964b8be360f443f714b2",
    strip_prefix = "rules_js-1.32.6",
    url = "https://github.com/aspect-build/rules_js/releases/download/v1.32.6/rules_js-v1.32.6.tar.gz",
)

http_archive(
    name = "aspect_rules_ts",
    sha256 = "8aabb2055629a7becae2e77ae828950d3581d7fc3602fe0276e6e039b65092cb",
    strip_prefix = "rules_ts-2.0.0",
    url = "https://github.com/aspect-build/rules_ts/releases/download/v2.0.0/rules_ts-v2.0.0.tar.gz",
)

http_archive(
    name = "aspect_rules_swc",
    sha256 = "8eb9e42ed166f20cacedfdb22d8d5b31156352eac190fc3347db55603745a2d8",
    strip_prefix = "rules_swc-1.1.0",
    url = "https://github.com/aspect-build/rules_swc/releases/download/v1.1.0/rules_swc-v1.1.0.tar.gz",
)

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "51dc53293afe317d2696d4d6433a4c33feedb7748a9e352072e2ec3c0dafd2c6",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.40.1/rules_go-v0.40.1.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.40.1/rules_go-v0.40.1.zip",
    ],
)

http_archive(
    name = "rules_proto_grpc",
    sha256 = "9ba7299c5eb6ec45b6b9a0ceb9916d0ab96789ac8218269322f0124c0c0d24e2",
    strip_prefix = "rules_proto_grpc-4.5.0",
    urls = ["https://github.com/rules-proto-grpc/rules_proto_grpc/releases/download/4.5.0/rules_proto_grpc-4.5.0.tar.gz"],
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
    sha256 = "6357de5982dd32526e02278221bb8d6aa45717ba9bbacf43686b130aa2c72e1e",
    urls = ["https://github.com/bazelbuild/rules_rust/releases/download/0.30.0/rules_rust-v0.30.0.tar.gz"],
)

# Container rules
http_archive(
    name = "rules_oci",
    sha256 = "c71c25ed333a4909d2dd77e0b16c39e9912525a98c7fa85144282be8d04ef54c",
    strip_prefix = "rules_oci-1.3.4",
    url = "https://github.com/bazel-contrib/rules_oci/releases/download/v1.3.4/rules_oci-v1.3.4.tar.gz",
)

http_archive(
    name = "rules_pkg",
    sha256 = "8c20f74bca25d2d442b327ae26768c02cf3c99e93fad0381f32be9aab1967675",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_pkg/releases/download/0.8.1/rules_pkg-0.8.1.tar.gz",
        "https://github.com/bazelbuild/rules_pkg/releases/download/0.8.1/rules_pkg-0.8.1.tar.gz",
    ],
)

http_archive(
    name = "container_structure_test",
    sha256 = "42edb647b51710cb917b5850380cc18a6c925ad195986f16e3b716887267a2d7",
    strip_prefix = "container-structure-test-104a53ede5f78fff72172639781ac52df9f5b18f",
    urls = ["https://github.com/GoogleContainerTools/container-structure-test/archive/104a53ede5f78fff72172639781ac52df9f5b18f.zip"],
)

# hermetic_cc_toolchain setup ================================
HERMETIC_CC_TOOLCHAIN_VERSION = "v2.1.2"

# Please note that we only use hermetic-cc for local development purpose and Nix, at it eases the path to cross-compile
# so we can produce container images locally on Mac laptops.
#
# @jhchabran See https://github.com/sourcegraph/sourcegraph/pull/55969, there is an ongoing issue with UBSAN
# and treesitter, that breaks the compilation of syntax-highlighter. Since we only use
# hermetic_cc for local development purposes, while it's a bit heavy handed for a --copt, it's acceptable
# at this point. Passing --copt=-fno-sanitize=undefined sadly doesn't fix the problem, which is why
# we have to patch to inject the flag.
http_archive(
    name = "hermetic_cc_toolchain",
    patch_args = ["-p1"],
    patches = [
        "//third_party/hermetic_cc:disable_ubsan.patch",
    ],
    sha256 = "28fc71b9b3191c312ee83faa1dc65b38eb70c3a57740368f7e7c7a49bedf3106",
    urls = [
        "https://mirror.bazel.build/github.com/uber/hermetic_cc_toolchain/releases/download/{0}/hermetic_cc_toolchain-{0}.tar.gz".format(HERMETIC_CC_TOOLCHAIN_VERSION),
        "https://github.com/uber/hermetic_cc_toolchain/releases/download/{0}/hermetic_cc_toolchain-{0}.tar.gz".format(HERMETIC_CC_TOOLCHAIN_VERSION),
    ],
)

# rules_js setup ================================
load("@aspect_rules_js//js:repositories.bzl", "rules_js_dependencies")

rules_js_dependencies()

# node toolchain setup ==========================
load("@rules_nodejs//nodejs:repositories.bzl", "nodejs_register_toolchains")

nodejs_register_toolchains(
    name = "nodejs",
    node_version = "20.8.0",
)

# rules_js npm setup ============================
load("@aspect_rules_js//npm:npm_import.bzl", "npm_translate_lock")

npm_translate_lock(
    name = "npm",
    npm_package_target_name = "{dirname}_pkg",
    npmrc = "//:.npmrc",
    pnpm_lock = "//:pnpm-lock.yaml",
    # Required for ESLint test targets.
    # See https://github.com/aspect-build/rules_js/issues/239
    # See `public-hoist-pattern[]=*eslint*` in the `.npmrc` of this monorepo.
    public_hoist_packages = {
        "@typescript-eslint/eslint-plugin": [""],
        "@typescript-eslint/parser@5.56.0_qxbo2xm47qt6fxnlmgbosp4hva": [""],
        "eslint-config-prettier": [""],
        "eslint-plugin-ban": [""],
        "eslint-plugin-etc": [""],
        "eslint-plugin-import": [""],
        "eslint-plugin-jest-dom": [""],
        "eslint-plugin-jsdoc": [""],
        "eslint-plugin-jsx-a11y": [""],
        "eslint-plugin-react@7.32.1_eslint_8.34.0": [""],
        "eslint-plugin-react-hooks": [""],
        "eslint-plugin-rxjs": [""],
        "eslint-plugin-unicorn": [""],
        "eslint-plugin-unused-imports": [""],
        "eslint-import-resolver-node": [""],
    },
    verify_node_modules_ignored = "//:.bazelignore",
)

# rules_ts npm setup ============================
load("@npm//:repositories.bzl", "npm_repositories")

npm_repositories()

load("@aspect_rules_ts//ts:repositories.bzl", "rules_ts_dependencies")

rules_ts_dependencies(ts_version = "4.9.5")

# rules_swc setup ==============================
load("@aspect_rules_swc//swc:dependencies.bzl", "rules_swc_dependencies")

rules_swc_dependencies()

load("@aspect_rules_swc//swc:repositories.bzl", "LATEST_SWC_VERSION", "swc_register_toolchains")

swc_register_toolchains(
    name = "swc",
    swc_version = LATEST_SWC_VERSION,
)

# rules_esbuild setup ===========================
http_archive(
    name = "aspect_rules_esbuild",
    sha256 = "84419868e43c714c0d909dca73039e2f25427fc04f352d2f4f7343ca33f60deb",
    strip_prefix = "rules_esbuild-0.15.3",
    url = "https://github.com/aspect-build/rules_esbuild/releases/download/v0.15.3/rules_esbuild-v0.15.3.tar.gz",
)

load("@aspect_rules_esbuild//esbuild:dependencies.bzl", "rules_esbuild_dependencies")

rules_esbuild_dependencies()

# Register a toolchain containing esbuild npm package and native bindings
load("@aspect_rules_esbuild//esbuild:repositories.bzl", "LATEST_ESBUILD_VERSION", "esbuild_register_toolchains")

esbuild_register_toolchains(
    name = "esbuild",
    esbuild_version = LATEST_ESBUILD_VERSION,
)

# Go toolchain setup
load("@rules_proto_grpc//:repositories.bzl", "rules_proto_grpc_repos", "rules_proto_grpc_toolchains")

rules_proto_grpc_toolchains()

rules_proto_grpc_repos()

load("@rules_proto_grpc//doc:repositories.bzl", rules_proto_grpc_doc_repos = "doc_repos")

rules_proto_grpc_doc_repos()

load("@rules_proto//proto:repositories.bzl", "rules_proto_dependencies", "rules_proto_toolchains")

rules_proto_dependencies()

rules_proto_toolchains()

load("@rules_buf//buf:repositories.bzl", "rules_buf_dependencies", "rules_buf_toolchains")

rules_buf_dependencies()

rules_buf_toolchains(version = "v1.11.0")

load("@rules_buf//gazelle/buf:repositories.bzl", "gazelle_buf_dependencies")

gazelle_buf_dependencies()

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

# Overrides the default provided protobuf dep from rules_go by a more
# recent one.
go_repository(
    name = "org_golang_google_protobuf",
    build_file_proto_mode = "disable_global",
    importpath = "google.golang.org/protobuf",
    sum = "h1:7QBf+IK2gx70Ap/hDsOmam3GE0v9HicjfEdAxE62UoM=",
    version = "v1.29.1",
)  # keep

# Pin protoc-gen-go-grpc to 1.3.0
# See also //:gen-go-grpc
go_repository(
    name = "org_golang_google_grpc_cmd_protoc_gen_go_grpc",
    build_file_proto_mode = "disable_global",
    importpath = "google.golang.org/grpc/cmd/protoc-gen-go-grpc",
    sum = "h1:rNBFJjBCOgVr9pWD7rs/knKL4FRTKgpZmsRfV214zcA=",
    version = "v1.3.0",
)  # keep

# gazelle:repository_macro deps.bzl%go_dependencies
go_dependencies()

go_rules_dependencies()

go_register_toolchains(
    nogo = "@//:sg_nogo",
    version = "1.20.10",
)

linter_dependencies()

gazelle_dependencies()

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

# rust toolchain setup
load("@rules_rust//rust:repositories.bzl", "rules_rust_dependencies", "rust_register_toolchains", "rust_repository_set")

rules_rust_dependencies()

rust_version = "1.73.0"

rust_register_toolchains(
    edition = "2021",
    # Keep in sync with syntax-highlighter/Dockerfile
    # and /syntax-highlighter/rust-toolchain.toml
    versions = [
        rust_version,
    ],
)

# Needed for locally cross-compiling rust binaries to linux/amd64 on a Mac laptop, when seeking to
# create container images in local for testing purposes.
rust_repository_set(
    name = "macos_arm_64",
    edition = "2021",
    exec_triple = "aarch64-apple-darwin",
    extra_target_triples = ["x86_64-unknown-linux-gnu"],
    versions = [rust_version],
)

load("@rules_rust//crate_universe:defs.bzl", "crates_repository")

crates_repository(
    name = "crate_index",
    cargo_config = "//syntax-highlighter:.cargo/config.toml",
    cargo_lockfile = "//syntax-highlighter:Cargo.lock",
    # this file has to be manually created and it will be filled when
    # the target is ran.
    # To regenerate this file run: CARGO_BAZEL_REPIN=1 bazel sync --only=crate_index
    lockfile = "//syntax-highlighter:Cargo.Bazel.lock",
    # glob doesn't work in WORKSPACE files: https://github.com/bazelbuild/bazel/issues/11935
    manifests = [
        "//syntax-highlighter:Cargo.toml",
        "//syntax-highlighter:crates/scip-macros/Cargo.toml",
        "//syntax-highlighter:crates/scip-syntax/Cargo.toml",
        "//syntax-highlighter:crates/scip-treesitter/Cargo.toml",
        "//syntax-highlighter:crates/scip-treesitter-languages/Cargo.toml",
        "//syntax-highlighter:crates/scip-treesitter-cli/Cargo.toml",
        "//syntax-highlighter:crates/sg-syntax/Cargo.toml",
    ],
)

load("@crate_index//:defs.bzl", "crate_repositories")

crate_repositories()

load("@hermetic_cc_toolchain//toolchain:defs.bzl", zig_toolchains = "toolchains")

zig_toolchains()

# containers steup       ===============================
load("@rules_oci//oci:dependencies.bzl", "rules_oci_dependencies")

rules_oci_dependencies()

load("@rules_oci//oci:repositories.bzl", "LATEST_CRANE_VERSION", "oci_register_toolchains")

oci_register_toolchains(
    name = "oci",
    crane_version = LATEST_CRANE_VERSION,
    # Uncommenting the zot toolchain will cause it to be used instead of crane for some tasks.
    # Note that it does not support docker-format images.
    # zot_version = LATEST_ZOT_VERSION,
)

# Optional, for oci_tarball rule
load("@rules_pkg//:deps.bzl", "rules_pkg_dependencies")

rules_pkg_dependencies()

load("//dev:oci_deps.bzl", "oci_deps")

oci_deps()

load("//cmd/embeddings/shared:assets.bzl", "embbedings_assets_deps")

embbedings_assets_deps()

load("@container_structure_test//:repositories.bzl", "container_structure_test_register_toolchain")

container_structure_test_register_toolchain(name = "cst")

load("//dev:tool_deps.bzl", "tool_deps")

tool_deps()

load("//tools/release:schema_deps.bzl", "schema_deps")

schema_deps()
