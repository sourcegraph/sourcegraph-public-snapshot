load("@npm//:mocha/package_json.bzl", mocha_bin = "bin")
load("@aspect_rules_esbuild//esbuild:defs.bzl", "esbuild")
load("@aspect_rules_js//js:defs.bzl", "js_run_binary")
load("@bazel_skylib//rules:build_test.bzl", "build_test")

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

    # Used by require.resolve
    "axe-core",
]

# ... some of which are needed at runtime
NON_BUNDLED_DEPS = [
    "//:node_modules/jsonc-parser",
    "//:node_modules/puppeteer",
    "//:node_modules/@axe-core/puppeteer",
    "//:node_modules/axe-core",
    "//client/web:node_modules/@sourcegraph/build-config",
]

def mocha_test(name, tests, deps = [], args = [], data = [], env = {}, is_percy_enabled = False, **kwargs):
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
        "$(rootpath //:mocha_config)",
        "'$(rootpath :%s)/**/*.test.js'" % bundle_name,
        "--retries 4",
    ] + args

    data = data + NON_BUNDLED_DEPS + [
        "//:mocha_config",
        ":%s" % bundle_name,
    ]

    # `--define` flags are used to set environment variables here because
    # we use `js_run_binary` as a target and it doesn't work with `--test_env`.
    env = dict(env, **{
        "HEADLESS": "$(E2E_HEADLESS)",
        # Add environment variable so that mocha writes its test xml
        # to the location Bazel expects.
        "MOCHA_FILE": "$$XML_OUTPUT_FILE",

        # TODO(bazel): e2e test environment
        "TEST_USER_EMAIL": "test@sourcegraph.com",
        "TEST_USER_PASSWORD": "supersecurepassword",
        "SOURCEGRAPH_BASE_URL": "$(E2E_SOURCEGRAPH_BASE_URL)",
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
        "DISPLAY": "$(DISPLAY)",
    })

    if is_percy_enabled:
        # Extract test specific arguments.
        flaky = kwargs.pop("flaky")
        timeout = kwargs.pop("timeout")

        binary_name = "%s_binary" % name

        # `js_run_binary` is used here in the combination with `build_test` instead of
        # `js_test` because only `js_run_binary` currntly supports the `stamp` attribute.
        # otherwise we could use js_binary with bazel test.
        # https://docs.aspect.build/rules/aspect_rules_js/docs/js_run_binary#stamp
        js_run_binary(
            name = binary_name,
            args = args,
            env = dict(env, **{
                "PERCY_ON": "true",
                "PERCY_TOKEN": "$(PERCY_TOKEN)",
            }),
            srcs = data,
            out_dirs = ["out"],
            silent_on_success = True,
            # Executed mocha tests with Percy enabled via `percy exec -- mocha ...`
            # Prepends volatile env variables to the command to make Percy aware of the
            # current git branch and commit.
            tool = "//client/shared/dev:run_mocha_tests_with_percy",
            testonly = True,
            **kwargs
        )

        build_test(
            name = name,
            targets = [binary_name],
            timeout = timeout,
            flaky = flaky,
        )
    else:
        mocha_bin.mocha_test(
            name = name,
            args = args,
            data = data,
            env = env,
            **kwargs
        )
