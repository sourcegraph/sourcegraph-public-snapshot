load("@npm//:@percy/cli/package_json.bzl", percy_bin = "bin")
load("@npm//:mocha/package_json.bzl", mocha_bin = "bin")
load("@aspect_rules_esbuild//esbuild:defs.bzl", "esbuild")
load("@aspect_bazel_lib//lib:stamping.bzl", "STAMP_ATTRS", "maybe_stamp")
load("@bazel_skylib//rules:build_test.bzl", "build_test")
load("//dev:integration.bzl", "run_integration")

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
    "/Users/val/Desktop/sourcegraph-root/sourcegraph/client/testing/src/percySnapshotToDisk2.js",
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
        # ".mocharc.js",
        "$(rootpath //:mocha_config)",
        # "$(location //:mocha_config)",
        "$(rootpath :%s)/**/*.test.js" % bundle_name,
        # "$(execpath :%s)/**/*.test.js" % bundle_name,
        # "$(location :%s)/**/*.test.js" % bundle_name,
        "--retries 0",
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
        "JS_BINARY__LOG_DEBUG": "true",
    })

    if is_percy_enabled:
        percy_args = [
            "exec",
            "--quiet",
            "--",
            # TODO: figure out how to get this path from "//:node_modules/mocha"
            "node_modules/mocha/bin/mocha",
        ]

        run_integration(
            name = name,
            # stamp = 1,  # uses BUILDKITE_BRANCH and BUILDKITE_COMMIT in Percy report
            args = args,
            data = data + args_data + NON_BUNDLED_DEPS + ["//:node_modules/mocha"],
            env = dict(env, **{
                "PERCY_ON": "true",
                "PERCY_TOKEN": "$$PERCY_TOKEN",
                # "PERCY_BROWSER_EXECUTABLE": "/private/var/tmp/_bazel_val/b74f98c07c1f7901aeeb711da373452f/execroot/__main__/bazel-out/darwin_arm64-fastbuild/bin/node_modules/.aspect_rules_js/puppeteer@13.7.0/node_modules/puppeteer/.local-chromium/mac-982053/chrome-mac/Chromium.app/Contents/MacOS/Chromium",
            }),
            **kwargs
        )

        # TODO: continue the work on the PercySnapshotToDisk idea.
        # exec_name = "%s_exec" % name

        # percy_bin.percy(
        #     name = exec_name,
        #     testonly = True,
        #     outs = ["snapshots.snap"],
        #     silent_on_success = False,
        #     use_execroot_entry_point = True,
        #     # stamp = 1,  # uses BUILDKITE_BRANCH and BUILDKITE_COMMIT in Percy report
        #     args = percy_args + args,
        #     # chdir = native.package_name(),
        #     srcs = data + args_data + NON_BUNDLED_DEPS + ["//:node_modules/mocha"],
        #     env = dict(env, **{
        #         "PERCY_ON": "true",
        #         "PERCY_TOKEN": "$$PERCY_TOKEN",
        #     }),
        #     tags = [
        #         "requires-network",
        #     ],
        # )

        # build_test(
        #     name = name,
        #     targets = [":%s" % exec_name],
        #     **kwargs
        # )

    else:
        mocha_bin.mocha_test(
            name = name,
            args = args,
            data = data + args_data + NON_BUNDLED_DEPS,
            env = env,
            **kwargs
        )
