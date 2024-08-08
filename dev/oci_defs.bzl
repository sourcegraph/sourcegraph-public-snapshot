"""OCI bazel defs"""

load("@rules_oci//oci:defs.bzl", _oci_image = "oci_image", _oci_load = "oci_load", _oci_push = "oci_push")
load("@rules_pkg//:pkg.bzl", _pkg_tar = "pkg_tar")

REGISTRY_REPOSITORY_PREFIX = "europe-west1-docker.pkg.dev/sourcegraph-security-logging/rules-oci-test/{}"
# REGISTRY_REPOSITORY_PREFIX = "us.gcr.io/sourcegraph-dev/{}"

# Passthrough the @rules_oci oci_push, so users only have to import this file and not @rules_oci//oci:defs.bzl
oci_push = _oci_push

def image_repository(image):
    return REGISTRY_REPOSITORY_PREFIX.format(image)

def oci_tarball(name, **kwargs):
    _oci_load(
        name = name,
        # Don't build this by default with bazel build //... since most oci_tarball
        # targets do not need to be built on CI. This prevents the remote cache from
        # being overwhelmed in the event that oci_tarballs are cache busted en masse.
        tags = kwargs.pop("tags", []) + ["manual"],
        **kwargs
    )

# Apply a transition on oci_image targets and their deps to apply a transition on platforms
# to build binaries for Linux when building on MacOS.
def oci_image(name, **kwargs):
    _oci_image(
        name = name + "_underlying",
        tars = kwargs.pop("tars", []) + ["//internal/version:stamps"],
        **kwargs
    )

    oci_image_cross(
        name = name,
        image = ":" + name + "_underlying",
        platform = select({
            "@platforms//os:macos": Label("@zig_sdk//platform:linux_amd64"),
            "//conditions:default": Label("@platforms//host"),
        }),
        visibility = kwargs.pop("visibility", ["//visibility:public"]),
    )

def _oci_image_cross_impl(ctx):
    runfiles = ctx.runfiles(files = ctx.files.image)
    runfiles = runfiles.merge(ctx.attr.image[0][DefaultInfo].default_runfiles)
    return [
        DefaultInfo(
            files = depset(ctx.files.image),
            runfiles = runfiles,
        ),
    ]

# rule that allows transitioning in order to transition an oci_image target and its deps
oci_image_cross = rule(
    implementation = _oci_image_cross_impl,
    attrs = {
        "image": attr.label(cfg = transition(
            implementation = lambda settings, attr: [
                {"//command_line_option:platforms": str(attr.platform), "//command_line_option:compilation_mode": "opt"},
            ],
            inputs = [],
            outputs = ["//command_line_option:platforms", "//command_line_option:compilation_mode"],
        )),
        "platform": attr.label(),
        "_allowlist_function_transition": attr.label(
            default = "@bazel_tools//tools/allowlists/function_transition_allowlist",
        ),
    },
)

def pkg_tar(name, **kwargs):
    _pkg_tar(
        name = name,
        extension = "tar.gz",
        **kwargs
    )
