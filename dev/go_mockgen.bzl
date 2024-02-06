load("@aspect_bazel_lib//lib:run_binary.bzl", "run_binary")
load("@io_bazel_rules_go//go:def.bzl", "GoArchive")
load("@bazel_skylib//lib:paths.bzl", "paths")

def go_mockgen1(name, out_file, pkg, file_prefix):
    run_binary(
        name = "mockgen-{}".format(name),
        mnemonic = "GoMockgen",
        tool = "@com_github_derision_test_go_mockgen//cmd/go-mockgen:go-mockgen",
        args = ["--filename", out_file, "--package", pkg, "--file-prefix", file_prefix],
        outs = [out_file],
    )

    # TODO: copy to source tree

def _go_mockgen(ctx):
    print(ctx.attr.deps[GoArchive].data.file.path)
    print(ctx.attr.deps[GoArchive].data.importmap)
    print(ctx.attr.deps[GoArchive].data.importpath)
    print(ctx.attr.gomockgen.files.to_list())
    print(paths.dirname(ctx.attr.out), ctx.attr.out)

    dst = ctx.actions.declare_directory(paths.dirname(ctx.attr.out))

    ctx.actions.run(
        mnemonic = "GoMockgen",
        arguments = [
            "--package",  # output package name
            paths.basename(paths.dirname(ctx.attr.out)),
            "--import-path",
            ctx.attr.deps[GoArchive].data.importpath,
            "--interfaces",
            ctx.attr.interfaces[0],
            "--filename",
            ctx.attr.out,
            "--force",
            "--disable-formatting",
            "--for-test",
            "--archives",
            "{}={}={}={}".format(
                ctx.attr.deps[GoArchive].data.importpath,
                ctx.attr.deps[GoArchive].data.importmap,
                ctx.attr.deps[GoArchive].data.file.path,
                ctx.attr.deps[GoArchive].data.file.path,
            ),
            "--sources",
            ctx.attr.deps[GoArchive].data.srcs[0],
            # TODO: multiple archives, multiple sources, and the "path" positional args
        ],
        executable = ctx.executable.gomockgen,
        outputs = [ctx.outputs.out],
        progress_message = "Running go-mockgen to generate %s" % ctx.outputs.out.short_path,
    )

    return [
        DefaultInfo(
            files = depset([dst]),
        ),
    ]

go_mockgen = rule(
    implementation = _go_mockgen,
    attrs = {
        "interfaces": attr.string_list(
            mandator = True,
        ),
        "deps": attr.label(
            # change to label_list?
            providers = [GoArchive],
            allow_empty = False,
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
