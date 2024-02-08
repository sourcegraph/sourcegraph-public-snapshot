load("@io_bazel_rules_go//go:def.bzl", "GoArchive", "GoSource")
load("@aspect_bazel_lib//lib:yq.bzl", "yq")

def _go_mockgen_run(ctx):
    dst = ctx.actions.declare_file(ctx.attr.out)
    # config_dst = ctx.actions.declare_file("mockgen.yaml")

    print("IMPORT PATH", "github.com/sourcegraph/sourcegraph/" + ctx.label.package)

    args = [
        "--import-path",
        "github.com/sourcegraph/sourcegraph/" + ctx.label.package,
        "--filename",
        dst.path,
        "--force",
        "--goimports",
        ctx.executable._goimports.path,
        "--stdlibroot",
        "{path}/{os}_{arch}".format(
            path = ctx.attr._go_stdlib[GoSource].stdlib.libs[0].path,
            os = ctx.attr._go_stdlib[GoSource].mode.goos,
            arch = ctx.attr._go_stdlib[GoSource].mode.goarch,
        ),
    ]

    for interface in ctx.attr.interfaces:
        args.extend([
            "--interfaces",
            interface,
        ])

    for dep in ctx.attr.deps:
        args.extend([
            "--archives",
            "{}={}={}={}".format(
                dep[GoArchive].data.importpath,
                dep[GoArchive].data.importmap,
                dep[GoArchive].data.file.path,
                dep[GoArchive].data.file.path,
            ),
        ])

    if ctx.attr.outpackage != None:
        args.extend([
            "--package",  # will be inferred if not explicitly passed
            ctx.attr.outpackage,
        ])

    # transformer_args = [config_dst.path]
    transformer_args = []
    deps = []
    for dep in ctx.attr.deps:
        for a in dep[GoArchive].direct:
            transformer_args.append("--archives")
            transformer_args.append("{}={}={}={}".format(
                a.data.importpath,
                a.data.importmap,
                a.data.file.path,
                a.data.file.path,
            ))
            deps.append(depset(direct = [a.data.file]))

    for dep in ctx.attr.deps:
        for src in dep[GoArchive].data.srcs:
            transformer_args.extend(["--source-files", "%s=%s" % (dep[GoArchive].data.importpath, src.path)])
            deps.append(depset(direct = [src]))

    args.extend(transformer_args)  # TODO: remove when moving to transformer
    for dep in ctx.attr.deps:
        args.append(dep[GoArchive].data.importpath)

    manifests = []
    for f in ctx.attr.manifests:
        manifests.extend(f.files.to_list())

    # ctx.actions.run(
    #     mnemonic = "GoMockgenConfigTransform",
    #     executable = ctx.executable._gomockgen_transformer,
    #     args = [config_dst.path],
    #     outputs = [config_dst],
    #     inputs = depset(direct = manifests),
    #     progress_message = "Transforming go-mockgen config for %s" % ctx.label.package,  # name = path? double check
    # )

    action_direct_deps = [ctx.attr._go_stdlib[GoSource].stdlib.libs[0]]
    for dep in ctx.attr.deps:
        action_direct_deps.append(dep[GoArchive].data.file)

    ctx.actions.run(
        mnemonic = "GoMockgen",
        arguments = args,
        executable = ctx.executable._gomockgen,
        outputs = [dst],
        tools = [ctx.executable._goimports],
        inputs = depset(
            # TODO: config file when moving to yaml
            direct = action_direct_deps,
            transitive = deps,
        ),
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
        "outpackage": attr.string(
            mandatory = False,
        ),
        "deps": attr.label_list(
            # change to label_list?
            providers = [GoArchive],
            allow_empty = False,
            mandatory = True,
        ),
        "out": attr.string(
            mandatory = True,
        ),
        "manifests": attr.label_list(
            allow_files = True,
            mandatory = False,
        ),
        "_gomockgen_transformer": attr.label(
            default = Label("//dev/go-mockgen-transformer:go-mockgen-transformer"),
            executable = True,
            cfg = "exec",
        ),
        "_gomockgen": attr.label(
            default = Label("@com_github_derision_test_go_mockgen//cmd/go-mockgen:go-mockgen"),
            executable = True,
            cfg = "exec",
        ),
        "_goimports": attr.label(
            default = Label("@org_golang_x_tools//cmd/goimports:goimports"),
            executable = True,
            cfg = "exec",
        ),
        "_go_stdlib": attr.label(
            providers = [GoSource],
            default = Label("@io_bazel_rules_go//:stdlib"),
        ),
        "_runfiles_lib": attr.label(
            default = "@bazel_tools//tools/bash/runfiles",
            allow_single_file = True,
        ),
    },
)
