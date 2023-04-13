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

    # UMD modules
    "jsonc-parser",

    # Dependencies with bundling issues
    "@sourcegraph/build-config",
]

# ... some of which are needed at runtime
NON_BUNDLED_DEPS = [
    "//:node_modules/jsonc-parser",
    "//:node_modules/puppeteer",
    "//:node_modules/@axe-core/puppeteer",
    "//:node_modules/axe-core",
    "//client/web:node_modules/@sourcegraph/build-config",
]

def mocha_test(name, tests, deps = [], args = [], data = [], env = {}, **kwargs):
    bundle_name = "%s_bundle" % name

    # Bundle the tests to remove the use of esm modules in tests
    esbuild(
        name = bundle_name,
        testonly = True,
        entry_points = tests,
        platform = "node",
        target = "esnext",
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

    args = [
        "--config",
        "$(location //:mocha_config)",
        "$(location :%s)/**/*.test.js" % bundle_name,
        "--retries 4",
    ] + args

    args_data = [
        "//:mocha_config",
        ":%s" % bundle_name,
    ]

    env = dict(env, **{
        "HEADLESS": "$$E2E_HEADLESS",
        # Add environment variable so that mocha writes its test xml
        # to the location Bazel expects.
        "MOCHA_FILE": "$$XML_OUTPUT_FILE",

        # TODO(bazel): e2e test environment
        "TEST_USER_EMAIL": "test@sourcegraph.com",
        "TEST_USER_PASSWORD": "supersecurepassword",
        "SOURCEGRAPH_BASE_URL": "$$E2E_SOURCEGRAPH_BASE_URL",
        "GH_TOKEN": "$$GH_TOKEN",
        "SOURCEGRAPH_SUDO_TOKEN": "fake-sg-token",
        "NO_CLEANUP": "false",
        "KEEP_BROWSER": "false",
        "DEVTOOLS": "false",
        "BROWSER": "chrome",
        "WINDOW_WIDTH": "1280",
        "WINDOW_HEIGHT": "1024",
        "LOG_BROWSER_CONSOLE": "false",

        # Enable findDom on CodeMirror
        "INTEGRATION_TESTS": "true",

        # Puppeteer config
        "DISPLAY": ":99",
    })

    bin.mocha_test(
        name = name,
        args = args,
        data = data + args_data + NON_BUNDLED_DEPS,
        env = env,
        **kwargs
    )
