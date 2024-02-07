load("@bazel_skylib//lib:paths.bzl", "paths")
load("@io_bazel_rules_go//go:def.bzl", "GoArchive", "GoSource")

#     # TODO: copy to source tree

def _go_mockgen_run(ctx):
    print("archive file", ctx.attr.deps[GoArchive].data.file.path)
    print("importmap", ctx.attr.deps[GoArchive].data.importmap)
    print("importpath", ctx.attr.deps[GoArchive].data.importpath)
    print("gomockgen files", ctx.attr.gomockgen.files.to_list())
    print("output dirname", paths.dirname(ctx.attr.out), "full output", ctx.attr.out)
    print("sources", ctx.attr.deps[GoArchive].data.srcs)
    print("direct deps", ctx.attr.deps[GoArchive].direct[0].data.file.path)

    dst = ctx.actions.declare_file(ctx.attr.out)
    print("destination", dst.path)

    print("stdlib", ctx.attr._go_stdlib[GoSource].stdlib.libs[0].path)

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
        "--for-test",
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
        progress_message = "Running go-mockgen to generate %s" % dst.path,
    )

    return [
        DefaultInfo(
            files = depset([dst]),
        ),
    ]

go_mockgen = rule(
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

def _go_mockgen_config(rctx):
    yq_path = rctx.path(
        Label("@yq-{}-{}//file:downloaded".format({
            "mac os x": "darwin",
            "linux": "linux",
        }[rctx.os.name], {
            "aarch64": "arm64",
            "arm64": "arm64",
            "amd64": "amd64",
            "x86_64": "amd64",
            "x86": "amd64",
        }[rctx.os.arch])),
    )

    mockgen_path = rctx.path(Label("//:mockgen.yaml"))

    # TODO: we have to write the targets to a .bzl file that is NOT BUILD.bazel,
    # so that we can load the file and run a function to add the targets to the
    # right namespace. Else they'd be trying to reference e.g. //cmd/gitserver/internal
    # in the go-mockgen bazel repository workspace instead of the sourcegraph workspace
    # rctx.file("BUILD", "[]")
    rctx.template("BUILD", Label("//dev:go-mockgen.tpl"))

    base_config = yaml_to_json(rctx, yq_path, mockgen_path)

    # prefix for the top of each generated file
    file_prefix = base_config["file-prefix"]

    for config in base_config["include-config-paths"]:
        # parse each to determine what to generate and where
        c = yaml_to_json(rctx, yq_path, rctx.path(Label("//:{}".format(config))))

    # TODO: finish the owl

def yaml_to_json(rctx, yq_path, path):
    res = rctx.execute([yq_path, "-p", "yaml", "-o", "json", path])
    if res.return_code != 0:
        fail("failed to run yq: {}".format(res.stderr))

    return json.decode(res.stdout)

go_mockgen_config = repository_rule(
    implementation = _go_mockgen_config,
)
