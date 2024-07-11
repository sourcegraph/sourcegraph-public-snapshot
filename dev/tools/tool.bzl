load("@aspect_bazel_lib//lib:output_files.bzl", "make_output_files")

# Convenience wrapper for rules that provide a single executable (binary/script/etc) that
# we want to rename. For example, files produced by http_file have the name `downloaded`,
# we may want to rename them with sh_binary, however this has a weird quirk whereby both
# the sh_binary wrapper "script" and the input binary/script itself are included in the outputs.
# This results in some less ergonomic usage when trying to use them in e.g. go tests, having to
# use $(rlocationpaths ...) and then filepath.Dir(runfiles.Rlocation(strings.Split(..., " ")[0])) in
# order to get the path.
# With this macro, its slightly simplified by being able to use $(rlocationpath ...) (singular) and
# not having to strings.Split(..., " ") as a result.
# See more: https://github.com/bazelbuild/bazel/issues/11820
def tool(name, src, visibility = ["//visibility:public"]):
    native.sh_binary(
        name = name + "_sh",
        srcs = src,
    )
    make_bin_and_deps_available(
        name = name,
        out = name,
        data = make_output_files(
            name = name + "_out",
            target = name + "_sh",
            paths = [native.package_name() + "/" + name + "_sh"],
        ),
        visibility = visibility,
    )

def _make_bin_and_deps_available_impl(ctx):
    symlink = ctx.actions.declare_file(ctx.attr.out)
    ctx.actions.symlink(output = symlink, target_file = ctx.file.data)
    return [
        DefaultInfo(
            executable = symlink,
            files = depset(direct = [symlink]),
            runfiles = ctx.runfiles(files = ctx.files.data),
        ),
    ]

make_bin_and_deps_available = rule(
    _make_bin_and_deps_available_impl,
    executable = True,
    attrs = {
        "out": attr.string(mandatory = True),
        "data": attr.label(mandatory = True, allow_single_file = True),
    },
)
