load("@aspect_bazel_lib//lib:copy_to_bin.bzl", "copy_files_to_bin_actions")
load("//dev:js_lib.bzl", "gather_files_from_js_providers", "gather_runfiles")
load("@aspect_rules_js//js:defs.bzl", "js_library")
load("@aspect_rules_js//js:providers.bzl", "JsInfo")
load("@bazel_skylib//rules:build_test.bzl", "build_test")

def _custom_eslint_impl(ctx):
    copied_srcs = copy_files_to_bin_actions(ctx, ctx.files.srcs)

    inputs_depset = depset(
        copied_srcs,
        transitive = [gather_files_from_js_providers(
            targets = [ctx.attr.config] + ctx.attr.deps,
            include_sources = False,
            include_transitive_sources = False,
            include_declarations = True,
            include_npm_linked_packages = True,
        )],
    )

    runfiles = gather_runfiles(
        ctx = ctx,
        sources = [],
        data = [ctx.attr.config],
        deps = [],
    )

    args = ctx.actions.args()

    # TODO: add context on why it doesn't work well with `overrides` globs.
    # args.add("--no-eslintrc")
    # args.add_all(["--config", get_path(ctx.files.config[0])])

    args.add("--quiet")
    args.add_all(["--format", "./{}".format(ctx.files.formatter[0].short_path)])

    output = ctx.actions.declare_file(ctx.attr.output)
    # args.add_all(["--output-file", output.short_path])

    args.add_all([s.short_path for s in copied_srcs])
    # print("ARGS", args)

    env = {
        "BAZEL_BINDIR": ctx.bin_dir.path,
        # "JS_BINARY__LOG_DEBUG": "1",
        # "JS_BINARY__LOG_INFO": "1",
        # "JS_BINARY__LOG_ERROR": "1",
        # "JS_BINARY__SILENT_ON_SUCCESS": "0",
        "JS_BINARY__STDOUT_OUTPUT_FILE": output.path,
        "JS_BINARY__STDERR_OUTPUT_FILE": output.path,
        # "JS_BINARY__EXPECTED_EXIT_CODE": "0",
        # "JS_BINARY__EXIT_CODE_OUTPUT_FILE": output.path,
    }

    ctx.actions.run(
        env = env,
        inputs = inputs_depset,
        outputs = [output],
        executable = ctx.executable.binary,
        arguments = [args],
        mnemonic = "ESLint",
    )

    return [
        DefaultInfo(
            files = depset([output]),
            runfiles = runfiles,
        ),
        OutputGroupInfo(
            output = depset([output]),
            runfiles = runfiles.files,
        ),
    ]

_eslint_test_with_types = rule(
    implementation = _custom_eslint_impl,
    attrs = {
        "srcs": attr.label_list(allow_files = True),
        "deps": attr.label_list(default = [], providers = [JsInfo]),
        "config": attr.label(allow_single_file = True),
        "formatter": attr.label(allow_single_file = True, default = Label("//:eslint-relative-formatter")),
        "binary": attr.label(executable = True, cfg = "exec", allow_files = True),
        "output": attr.string(),
    },
)

def eslint_test_with_types(name, **kwargs):
    lint_name = "%s_lint" % name

    build_test(name, targets = [lint_name])

    _eslint_test_with_types(
        name = lint_name,
        output = "%s-output.txt" % name,
        **kwargs
    )

def eslint_config():
    client_package_path = "/".join(native.package_name().split("/")[:2])

    js_library(
        name = "eslint_config",
        testonly = True,
        srcs = [".eslintrc.js"],
        data = [
            ".eslintignore",
            "package.json",
            ":tsconfig",
        ],
        deps = [
            "//:eslint_config",
        ],
        visibility = ["//{}:__subpackages__".format(client_package_path)],
    )
