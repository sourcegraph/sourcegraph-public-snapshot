load("@bazel_skylib//lib:paths.bzl", "paths")
load("@io_bazel_rules_go//go:def.bzl", "GoArchive", "GoSource")

def _go_mockgen_run(ctx):
    dst = ctx.actions.declare_file(ctx.attr.out)

    args = [
        "--package",  # output package name
        paths.basename(paths.dirname(ctx.attr.out)),
        "--import-path",
        ctx.attr.deps[GoArchive].data.importpath,
        "--interfaces",
        ctx.attr.interfaces[0],
        "--filename",
        dst.path,
        "--force",
        "--disable-formatting",
        "--stdlibroot",
        "{path}/{os}_{arch}".format(
            path = ctx.attr._go_stdlib[GoSource].stdlib.libs[0].path,
            os = ctx.attr._go_stdlib[GoSource].mode.goos,
            arch = ctx.attr._go_stdlib[GoSource].mode.goarch,
        ),
        "--archives",
        "{}={}={}={}".format(
            ctx.attr.deps[GoArchive].data.importpath,
            ctx.attr.deps[GoArchive].data.importmap,
            ctx.attr.deps[GoArchive].data.file.path,
            ctx.attr.deps[GoArchive].data.file.path,
        ),
    ]

    deps = []
    for a in ctx.attr.deps[GoArchive].direct:
        args.append("--archives")
        args.append("{}={}={}={}".format(
            a.data.importpath,
            a.data.importmap,
            a.data.file.path,
            a.data.file.path,
        ))
        deps.append(depset(direct = [a.data.file]))

    for src in ctx.attr.deps[GoArchive].data.srcs:
        args.extend(["--sources", src.path])
        deps.append(depset(direct = [src]))

    args.append(ctx.attr.deps[GoArchive].data.importpath)

    ctx.actions.run(
        mnemonic = "GoMockgen",
        arguments = args,
        executable = ctx.executable.gomockgen,
        outputs = [dst],
        inputs = depset(direct = [ctx.attr.deps[GoArchive].data.file, ctx.attr._go_stdlib[GoSource].stdlib.libs[0]], transitive = deps),
        progress_message = "Running go-mockgen to generate %s" % dst.short_path,
    )

    return [
        DefaultInfo(
            files = depset([dst]),
        ),
    ]

go_mockgen_generate = rule(
    implementation = _go_mockgen_run,
    attrs = {
        "interfaces": attr.string_list(
            mandatory = True,
        ),
        "deps": attr.label(
            # change to label_list?
            providers = [GoArchive],
            # allow_empty = False,
            mandatory = True,
        ),
        "out": attr.string(
            mandatory = True,
        ),
        "gomockgen": attr.label(
            default = Label("@com_github_derision_test_go_mockgen//cmd/go-mockgen:go-mockgen"),
            executable = True,
            cfg = "exec",
        ),
        "_go_stdlib": attr.label(
            providers = [GoSource],
            default = Label("@io_bazel_rules_go//:stdlib"),
        ),
        "_runfiles_lib": attr.label(default = "@bazel_tools//tools/bash/runfiles", allow_single_file = True),
    },
)
