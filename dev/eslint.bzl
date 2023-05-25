load("@aspect_bazel_lib//lib:copy_to_bin.bzl", "copy_files_to_bin_actions")
load("//dev:js_lib.bzl", "gather_files_from_js_providers", "gather_runfiles")
load("@aspect_rules_js//js:defs.bzl", "js_library")
load("@aspect_rules_js//js:providers.bzl", "JsInfo")
load("@bazel_skylib//rules:build_test.bzl", "build_test")

def get_client_package_path():
    # Used to reference the `eslint_config` target in the client package
    # We assume that eslint config files are located at `client/<package>`
    return "/".join(native.package_name().split("/")[:2])

def eslint_config(deps = []):
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
        ] + deps,
        visibility = ["//{}:__subpackages__".format(get_client_package_path())],
    )

def _custom_eslint_impl(ctx):
    copied_srcs = copy_files_to_bin_actions(ctx, ctx.files.srcs)

    inputs_depset = depset(
        copied_srcs,
        transitive = [gather_files_from_js_providers(
            targets = [ctx.attr.config] + ctx.attr.deps,
            include_sources = False,
            include_transitive_sources = False,
            # We have to include declarations because we need to lint the types.
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

    # Ignore warnings and fail only on errors.
    args.add("--quiet")

    # Use the custom formatter to ouput relative paths.
    args.add_all(["--format", "./{}".format(ctx.files.formatter[0].short_path)])

    # Specify the files to lint.
    args.add_all([s.short_path for s in copied_srcs])

    # Declare the output file for the eslint output.
    output = ctx.actions.declare_file(ctx.attr.output)
    # args.add_all(["--output-file", output.short_path])

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
