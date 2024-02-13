"""OCI bazel defs"""

load("@rules_oci//oci:defs.bzl", _oci_image = "oci_image", _oci_push = "oci_push", _oci_tarball = "oci_tarball")
load("@with_cfg.bzl", "with_cfg")

REGISTRY_REPOSITORY_PREFIX = "europe-west1-docker.pkg.dev/sourcegraph-security-logging/rules-oci-test/{}"
# REGISTRY_REPOSITORY_PREFIX = "us.gcr.io/sourcegraph-dev/{}"

def image_repository(image):
    return REGISTRY_REPOSITORY_PREFIX.format(image)

def oci_tarball(name, **kwargs):
    _oci_tarball(
        name = name,
        # Don't build this by default with bazel build //... since most oci_tarball
        # targets do not need to be built on CI. This prevents the remote cache from
        # being overwhelmed in the event that oci_tarballs are cache busted en masse.
        tags = kwargs.pop("tags", []) + ["manual"],
        **kwargs
    )

_oci_image_builder = with_cfg(_oci_image)
_oci_image_builder.set("platforms", select({
    "@platforms//os:macos": [Label("@zig_sdk//platform:linux_amd64")],
    "//conditions:default": [],
}))

oci_image, _oci_image_internal = _oci_image_builder.build()
oci_push = _oci_push
