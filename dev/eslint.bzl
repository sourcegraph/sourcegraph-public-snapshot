load("@aspect_bazel_lib//lib:copy_to_bin.bzl", "copy_files_to_bin_actions")
load("//dev:js_lib.bzl", "gather_files_from_js_providers", "gather_runfiles")
load("@aspect_rules_js//js:defs.bzl", "js_library")
load("@aspect_rules_js//js:providers.bzl", "JsInfo")

def eslint_config_and_lint_root(name = "eslint_config", config_deps = [], root_js_deps = []):
    """
    Creates an ESLint configuration target and an ESLint test target for a client package root JS files.

    Args:
        name: The name of the ESLint configuration target.
        config_deps: A list of dependencies for the ESLint config target.
        root_js_deps: A list of dependencies for the `root_js_eslint` target.

    The macro assumes the presence of specific files (".eslintrc.js", ".eslintignore", "package.json")
    and a tsconfig target in the current directory. It adds a reference to a top-level ESLint configuration
    as a dependency, and sets the visibility to the current package and its subpackages.

    For the 'root_js_eslint' target, it assumes all '.js' files in the current directory as its sources and
    additional dependencies provided by 'root_js_deps'. It uses a global ESLint binary and the generated
    ESLint configuration as its config.

    Example usage:
        eslint_config_and_lint_root(
            config_deps = ["//my:dependency"],
            root_js_deps = ["//other:dependency"],
        )
    """

    js_library(
        name = name,
        testonly = True,
        srcs = ["//:eslint_config"],
        data = [
            "package.json",
            ":tsconfig",
        ],
        deps = [
            "//:eslint_config",
        ] + config_deps,
        visibility = ["//{}:__subpackages__".format(get_client_package_path())],
    )

    eslint_test_with_types(
        name = "root_js_eslint",
        srcs = native.glob(["*.js"]),
        config = ":eslint_config",
        deps = [
            "//:jest_config",  # required for import/extensions rule not to fail on the `jest.config.base` import.
            "//:node_modules/@types/node",
        ] + root_js_deps,
    )

# This private rule implementation wraps the ESLint binary.
# It executes ESLint against the provided source files and
# ensures that depenencies' type are available at lint time.
def _custom_eslint_impl(ctx):
    copied_srcs = copy_files_to_bin_actions(ctx, ctx.files.srcs)

    inputs_depset = depset(
        copied_srcs + [ctx.executable.binary],
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

    # Declare the output file for the ESLint output.
    report = ctx.actions.declare_file(ctx.attr.report)

    args = ctx.actions.args()  # Create the argument list for the ESLint command.
    args.add("--quiet")  # Ignore warnings and fail only on errors.
    args.add_all(["--format", "./{}".format(ctx.files.formatter[0].short_path)])  # Use the custom formatter to ouput relative paths.
    args.add_all([s.short_path for s in copied_srcs])  # Specify the files to lint.
    args.add_all(["--output-file", report.short_path])  # Specify the output file for the ESLint output.

    # Declare the output file for the exit code output.
    exit_code_out = ctx.actions.declare_file("exit_%s" % ctx.attr.report)

    env = {
        "BAZEL_BINDIR": ctx.bin_dir.path,
        # "JS_BINARY__LOG_DEBUG": "1",
        # "JS_BINARY__LOG_INFO": "1",
        # "JS_BINARY__LOG_ERROR": "1",
        # "JS_BINARY__SILENT_ON_SUCCESS": "0",
        # "JS_BINARY__STDOUT_OUTPUT_FILE": report.path,
        # "JS_BINARY__STDERR_OUTPUT_FILE": report.path,
        "JS_BINARY__EXIT_CODE_OUTPUT_FILE": exit_code_out.path,
    }

    # The script wrapper around the ESLint binary is essential to create an empty 'report'
    # file in cases where ESLint finds no errors. Bazel expects all declared outputs of
    # ctx.actions.run_shell to be created during its execution. Failure to do so results
    # in Bazel errors, hence if ESLint doesn't generate a 'report', we manually create one.
    command = """
        #!/usr/bin/env bash
        set -o pipefail -o errexit -o nounset

        # Call the ESLint @aspect_rules_js wrapper.
        "{binary}" "$@"

        # If the ESLint report is not created, create the empty one.
        if [ ! -f "{report}" ]; then
            touch "{report}"
        fi
    """.format(binary = ctx.executable.binary.path, report = report.path)

    # Generate and run a bash script to wrap the binary
    ctx.actions.run_shell(
        env = env,
        inputs = inputs_depset,
        outputs = [report, exit_code_out],
        command = command,
        arguments = [args],
        mnemonic = "ESLint",
        tools = ctx.attr.binary[DefaultInfo].default_runfiles.files,
    )

    return [
        DefaultInfo(
            files = depset([report]),
            runfiles = runfiles,
        ),
        OutputGroupInfo(
            report = depset([report]),
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
        "report": attr.string(),
    },
)

def eslint_test_with_types(name, **kwargs):
    """
    A higher-level function to perform an ESLint test on TypeScript files with type checking.

    Args:
        name: A string representing the name of the test.
        **kwargs: Arbitrary keyword arguments for additional customization of the test. This can
                  include the source files (`srcs`), dependencies (`deps`), ESLint configuration
                  (`config`), and more.

    This macro wraps the `_eslint_test_with_types` rule and subsequently runs a shell test to
    verify the output. It generates an output report named '<name>-output.txt' and a linting
    target with the name of the original test suffixed with '_lint'.

    Example usage:
        eslint_test_with_types(
            name = "my_test",
            srcs = ["my_file.ts"],
            deps = [":my_dependency"],
            testonly = True,
            config = ":my_eslint_config",
        )
    """
    lint_name = "%s_lint" % name
    report = "%s-output.txt" % name

    _eslint_test_with_types(
        testonly = True,
        name = lint_name,
        report = report,
        binary = "//:eslint",
        **kwargs
    )

    lint_target_name = ":%s" % lint_name

    native.sh_test(
        name = name,
        srcs = ["//dev:eslint-report-test.sh"],
        args = ["$(location %s)" % lint_target_name],
        data = [lint_target_name],
        timeout = "short",
    )

# This function provides the path to the client package, assuming
# that eslint config files are located at `client/<package>`.
def get_client_package_path():
    return "/".join(native.package_name().split("/")[:2])
