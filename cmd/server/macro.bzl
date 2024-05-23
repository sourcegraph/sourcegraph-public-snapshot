load("@rules_pkg//:pkg.bzl", "pkg_tar")
load("@io_bazel_rules_go//go:def.bzl", "GoArchive", "go_binary")

def get_last_segment(path):
    segments = path.split("/")
    last_segment = segments[-1]

    s = last_segment.split(":")
    if len(s) == 1:
        return last_segment
    else:
        return s[-1]

def container_dependencies(targets):
    for target in targets:
        name = get_last_segment(target)

        pkg_tar(
            name = "tar_{}".format(name),
            srcs = [target],
            remap_paths = {"/{}".format(name): "/usr/local/bin/{}".format(name)},
        )

def dependencies_tars(targets):
    tars = []
    for target in targets:
        name = get_last_segment(target)
        tars.append(":tar_{}".format(name))

    return tars

def go_binary_nobundle(name, **kwargs):
    go_binary(
        name = name + "_underlying",
        out = kwargs.pop("out", name + "_underlying"),
        **kwargs
    )

    go_binary_nobundle_rule(
        name = name,
        go_binary = ":" + name + "_underlying",
        visibility = kwargs.pop("visibility", ["//visibility:public"]),
    )

def _go_binary_nobundle_rule(ctx):
    # so that we can `bazel run` nobundle targets, we need to set `executable` in DefaultInfo.
    # But this can't be the output of a _different_ rule, it has to be the output of this rule.
    executable = ctx.actions.declare_file(ctx.executable.go_binary.basename)
    ctx.actions.symlink(output = executable, target_file = ctx.executable.go_binary)
    return [DefaultInfo(executable = executable, files = depset(ctx.files.go_binary))]

go_binary_nobundle_rule = rule(
    implementation = _go_binary_nobundle_rule,
    executable = True,
    attrs = {
        "go_binary": attr.label(
            providers = [GoArchive],
            executable = True,
            cfg = transition(
                implementation = lambda settings, attr: [{"//:integration_testing": True}],
                inputs = [],
                outputs = ["//:integration_testing"],
            ),
        ),
        "_allowlist_function_transition": attr.label(
            default = "@bazel_tools//tools/allowlists/function_transition_allowlist",
        ),
    },
)
