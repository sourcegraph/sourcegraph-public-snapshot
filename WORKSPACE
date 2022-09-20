load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "aspect_rules_js",
    sha256 = "fbfcfb5a1e4bccd8030b64522e561046731bdacd4a9ff8fa3334fed0e6c61af1",
    strip_prefix = "rules_js-9f6f33c72ac49dd829de7a622bd8fd81ec9d03e1",
    # strip_prefix = "rules_js-1.1.2",
    # url = "https://github.com/aspect-build/rules_js/archive/refs/tags/v1.1.2.tar.gz",
    url = "https://github.com/aspect-build/rules_js/archive/9f6f33c72ac49dd829de7a622bd8fd81ec9d03e1.tar.gz",
)

load("@aspect_rules_js//js:repositories.bzl", "rules_js_dependencies")

rules_js_dependencies()

load("@rules_nodejs//nodejs:repositories.bzl", "DEFAULT_NODE_VERSION", "nodejs_register_toolchains")

nodejs_register_toolchains(
    name = "nodejs",
    node_version = DEFAULT_NODE_VERSION,
)

load("@aspect_rules_js//npm:npm_import.bzl", "npm_translate_lock")

npm_translate_lock(
    name = "npm",
    # bins = {
    #     # derived from "bin" attribute in node_modules/typescript/package.json
    #     "typescript": {
    #         "tsc": "./bin/tsc",
    #         "tsserver": "./bin/tsserver",
    #     },
    # },
    pnpm_lock = "//:pnpm-lock.yaml",
    # package_json = "//:package.json",
    # yarn_lock = "//:yarn.lock",
    verify_node_modules_ignored = "//:.bazelignore",
)

load("@npm//:repositories.bzl", "npm_repositories")

npm_repositories()

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
http_archive(
    name = "aspect_rules_webpack",
    sha256 = "ec88e0c330661ed935534229c5e8480cb7ab281da4eca1b9b0b133374a724e66",
    strip_prefix = "rules_webpack-0.4.0",
    url = "https://github.com/aspect-build/rules_webpack/archive/refs/tags/v0.4.0.tar.gz",
)

load("@aspect_rules_webpack//webpack:dependencies.bzl", "rules_webpack_dependencies")

rules_webpack_dependencies()

load("@aspect_rules_webpack//webpack:repositories.bzl", "webpack_register_toolchains", "LATEST_VERSION")

webpack_register_toolchains(
    name = "webpack",
    webpack_version = LATEST_VERSION
)

http_archive(
    name = "aspect_rules_ts",
    sha256 = "3eb3171c26eb5d0951d51ae594695397218fb829e3798eea5557814723a1b3da",
    strip_prefix = "rules_ts-1.0.0-rc3",
    url = "https://github.com/aspect-build/rules_ts/archive/refs/tags/v1.0.0-rc3.tar.gz",
)

load("@aspect_rules_ts//ts:repositories.bzl", "LATEST_VERSION", "rules_ts_dependencies")

rules_ts_dependencies(ts_version = LATEST_VERSION)

http_archive(
    name = "aspect_rules_jest",
    sha256 = "bb3226707f9872185865a6381eb3a19311ca7b46e8ed475aad50975906a6cb6a",
    strip_prefix = "rules_jest-0.10.0",
    url = "https://github.com/aspect-build/rules_jest/archive/refs/tags/v0.10.0.tar.gz",
)

load("@aspect_rules_jest//jest:dependencies.bzl", "rules_jest_dependencies")

rules_jest_dependencies()

load("@aspect_rules_jest//jest:repositories.bzl", "LATEST_VERSION", "jest_repositories")

jest_repositories(
    name = "jest",
    jest_version = LATEST_VERSION,
)
