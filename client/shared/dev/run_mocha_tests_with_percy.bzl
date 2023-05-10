load("@bazel_skylib//rules:build_test.bzl", "build_test")
load("@aspect_rules_js//js:defs.bzl", "js_run_binary")

# TODO: document arguments and purpose
def run_mocha_tests_with_percy(name, args, data, env, timeout, flaky, **kwargs):
    js_run_binary(
        name = name,
        args = args,
        env = env,
        srcs = data,
        out_dirs = ["out"],
        silent_on_success = False,
        stamp = 1,  # uses BUILDKITE_BRANCH and BUILDKITE_COMMIT in Percy report
        tool = "//client/shared/dev:run_mocha_tests_with_percy",
        visibility = ["//visibility:public"],
        testonly = True,
        **kwargs
    )

    build_test(
        name = "%s_test" % name,
        targets = [name],
        visibility = ["//visibility:public"],
        timeout = timeout,
        flaky = flaky,
    )
