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
    sha256 = "f58d7be1bb0e4b7edb7a0085f969900345f5914e4e647b4f0d2650d5252aa87d",
    strip_prefix = "rules_js-1.8.0",
    url = "https://github.com/aspect-build/rules_js/archive/refs/tags/v1.8.0.tar.gz",
)

load("@aspect_rules_js//js:repositories.bzl", "rules_js_dependencies")

rules_js_dependencies()

http_archive(
    name = "rules_nodejs",
    sha256 = "50adf0b0ff6fc77d6909a790df02eefbbb3bc2b154ece3406361dda49607a7bd",
    urls = ["https://github.com/bazelbuild/rules_nodejs/releases/download/5.7.1/rules_nodejs-core-5.7.1.tar.gz"],
)

load("@rules_nodejs//nodejs:repositories.bzl", "DEFAULT_NODE_VERSION", "nodejs_register_toolchains")

nodejs_register_toolchains(
    name = "nodejs",
    node_version = DEFAULT_NODE_VERSION,
)

load("@aspect_rules_js//npm:npm_import.bzl", "npm_translate_lock")

npm_translate_lock(
    name = "npm",
    data = [
        # TODO: can remove these package.json labels after switching to a pnpm lockfile.
        "//:client/branded/package.json",
        "//:client/build-config/package.json",
        "//:client/client-api/package.json",
        "//:client/codeintellify/package.json",
        "//:client/common/package.json",
        "//:client/eslint-plugin-wildcard/package.json",
        "//:client/extension-api-types/package.json",
        "//:client/extension-api/package.json",
        "//:client/http-client/package.json",
        "//:client/jetbrains/package.json",
        "//:client/observability-client/package.json",
        "//:client/observability-server/package.json",
        "//:client/search-ui/package.json",
        "//:client/search/package.json",
        "//:client/shared/package.json",
        "//:client/storybook/package.json",
        "//:client/template-parser/package.json",
        "//:client/web/package.json",
        "//:client/wildcard/package.json",
        "//:pnpm-workspace.yaml",
    ],
    npmrc = "//:.npmrc",
    package_json = "//:package.json",  # TODO: not needed after switch to pnpm_lock
    verify_node_modules_ignored = "//:.bazelignore",
    yarn_lock = "//:yarn.lock",  # TODO: replace with pnpm_lock
)

load("@npm//:repositories.bzl", "npm_repositories")

npm_repositories()
