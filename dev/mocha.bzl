load("@npm//:mocha/package_json.bzl", "bin")
load("@aspect_rules_esbuild//esbuild:defs.bzl", "esbuild")

NON_BUNDLED = [
    # Dependencies loaded by mocha itself before the tests.
    # Should be outside the test bundle to ensure a single copy is used
    # mocha (launched by bazel) and the test bundle.
    "mocha",
    "puppeteer",

    # Dependencies used by mocha setup scripts.
    "abort-controller",
    "node-fetch",
    "console",

    # Dependencies with bundling issues
    "jsonc-parser"
]

# ... some of which are needed at runtime
NON_BUNDLED_DEPS = [
    "//:node_modules/jsonc-parser",
    "//:node_modules/puppeteer",
]

def mocha_test(name, tests, deps = [], args = [], data = [], env = {}, **kwargs):
    bundle_name = "%s_bundle" % name

    # Bundle the tests to remove the use of esm modules in tests
    esbuild(
        name = bundle_name,
        entry_points = tests,
        platform = "node",
        target = "node12",
        output_dir = True,
        external = NON_BUNDLED,
        sourcemap = "linked",
        deps = deps,
        config = {
            "loader": {
                # Packages such as fsevents require native .node binaries.
                ".node": "copy",
            },
        },
    )

    bin.mocha_test(
        name = name,
        args = [
            "--config",
            "$(location //:mocha_config)",
            "$(location :%s)/**/*.test.js" % bundle_name,
        ] + args,
        data = data + deps + [
            ":%s" % bundle_name,
            "//:mocha_config",
        ] + NON_BUNDLED_DEPS,
        env = dict(env, **{
            # Add environment variable so that mocha writes its test xml
            # to the location Bazel expects.
            "MOCHA_FILE": "$$XML_OUTPUT_FILE",

            # TODO(bazel): e2e test environment
            "GH_TOKEN": "fake-token",
            "TEST_USER_PASSWORD": "fake-password",
        }),
        tags = ["manual"],
        **kwargs
    )
