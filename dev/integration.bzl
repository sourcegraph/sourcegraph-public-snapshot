# load("@aspect_rules_ts//ts:defs.bzl", "ts_config")
load("//dev:defs.bzl", "ts_project")
load("@bazel_skylib//rules:build_test.bzl", "build_test")

# load("//dev:mocha.bzl", "mocha_test", "wow")
# load("@npm//:@percy/cli/package_json.bzl", percy_bin = "bin")
load("@aspect_rules_js//js:defs.bzl", "js_binary", "js_run_binary", "js_test")

# integration/ does not contain a src/
# gazelle:js_files **/*.{ts,tsx}

# gazelle:js_resolve sourcegraph //client/shared:node_modules/@sourcegraph/client-api

# ts_project(
#     name = "kek",
#     srcs = [
#         "kek.ts",
#     ],
#     module = "commonjs",
#     tsconfig = "//client/web/src/integration:tsconfig",
#     deps = [],
# )

# percy_bin.percy_binary(
#     name = "percy",
#     testonly = True,
#     visibility = ["//visibility:public"],
# )

# wow(name = "wow")

def run_integration(name, args, data, env, timeout, flaky, **kwargs):
    js_run_binary(
        name = name,
        args = args,
        srcs = data,
        env = env,
        # chdir = "../../../",  # TODO fix that!
        out_dirs = ["out"],
        silent_on_success = False,
        stamp = 1,
        tool = "//client/web/src/integration/kek:start-integration-tests",
        visibility = ["//visibility:public"],
        testonly = True,
        **kwargs
    )

    build_test(
        name = "kek_test",
        targets = [name],
        visibility = ["//visibility:public"],
        timeout = timeout,
        flaky = flaky,
    )
