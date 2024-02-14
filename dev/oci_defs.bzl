"""OCI bazel defs"""

load("@rules_oci//oci:defs.bzl", _oci_image = "oci_image", _oci_push = "oci_push", _oci_tarball = "oci_tarball")

REGISTRY_REPOSITORY_PREFIX = "europe-west1-docker.pkg.dev/sourcegraph-security-logging/rules-oci-test/{}"
# REGISTRY_REPOSITORY_PREFIX = "us.gcr.io/sourcegraph-dev/{}"

def image_repository(image):
    return REGISTRY_REPOSITORY_PREFIX.format(image)

def oci_tarball(name, **kwargs):
    _oci_tarball(
        name = name,
        # Don't build this by default with bazel build //... since
        # it is currently set to 'no-remote' in the root .bazelrc to reduce
        # remote cache traffic when oci_tarballs are cache busted en masse.
        tags = kwargs.pop("tags", []) + ["manual"],
        **kwargs
    )

oci_image = _oci_image
oci_push = _oci_push
