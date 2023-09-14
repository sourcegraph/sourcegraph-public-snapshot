load("@aspect_rules_webpack//webpack:defs.bzl", _webpack_bundle = "webpack_bundle", _webpack_devserver = "webpack_devserver")
load("@aspect_bazel_lib//lib:copy_to_directory.bzl", "copy_to_directory")

def webpack_bundle(name, **kwargs):
    _webpack_bundle(
        name = name,
        webpack = "//dev:webpack",
        **kwargs
    )

def webpack_web_app(name, **kwargs):
    bundle_name = "%s_bundle" % name

    _webpack_bundle(
        name = bundle_name,
        webpack = "//dev:webpack",
        **kwargs
    )

    copy_to_directory(
        name = name,
        # flatten static assets
        # https://docs.aspect.build/rules/aspect_bazel_lib/docs/copy_to_directory/#root_paths
        root_paths = ["ui/assets", "client/web/%s" % bundle_name],
        srcs = ["//ui/assets/img:img", ":%s" % bundle_name],
        visibility = ["//visibility:public"],
    )

def webpack_devserver(name, **kwargs):
    _webpack_devserver(
        name = name,
        webpack = "//dev:webpack",
        **kwargs
    )
