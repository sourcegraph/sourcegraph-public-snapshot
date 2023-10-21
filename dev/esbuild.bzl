load("@aspect_bazel_lib//lib:copy_to_directory.bzl", "copy_to_directory")
load("@aspect_rules_esbuild//esbuild:defs.bzl", "esbuild")

def esbuild_web_app(name, **kwargs):
    bundle_name = "%s_bundle" % name

    esbuild(
        name = bundle_name,
        **kwargs
    )

    copy_to_directory(
        name = name,
        # flatten static assets
        # https://docs.aspect.build/rules/aspect_bazel_lib/docs/copy_to_directory/#root_paths
        root_paths = ["client/web/dist", "client/web/%s" % bundle_name],
        srcs = ["//client/web/dist/img:img", ":%s" % bundle_name],
        visibility = ["//visibility:public"],
    )
