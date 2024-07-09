"""
Load external dependencies for base images
"""

load("@rules_oci//oci:pull.bzl", "oci_pull")

# Looking to update the base images?
#
# The update process has changed and no longer uses this file - for details see:
#     https://sourcegraph.notion.site/Add-and-Update-Base-Images-e809e5a79cb440d1a5d459c570a670f2#16838a6cbd9644bfaf3ff1b68ae4b595
#
# However, several legacy images are still updated via this file.

def oci_deps():
    """
    The image definitions and their digests
    """
    oci_pull(
        name = "scip-java",
        digest = "sha256:808b063b7376cfc0a4937d89ddc3d4dd9652d10609865fae3f3b34302132737a",
        image = "index.docker.io/sourcegraph/scip-java",
    )

    # The following image digests are from tag 252535_2023-11-28_5.2-82b5f4f5d73f. sg wolfi update-hashes DOES NOT update these digests.
    # To rebuild these legacy images using docker and outside of bazel you can either push a branch to:
    # - docker-images-candidates-notest/<your banch name here>
    # or you can run `sg ci build docker-images-candidates-notest`
    oci_pull(
        name = "legacy_alpine-3.14_base",
        digest = "sha256:581afabd476b4918b14295ae6dd184f4a3783c64bab8bde9ad7b11ea984498a8",
        image = "index.docker.io/sourcegraph/alpine-3.14",
    )

    # https://hub.docker.com/_/docker/tags?name=dind
    # Tag: docker:27.0.3-dind
    oci_pull(
        name = "upstream_dind_base",
        digest = "sha256:2632da0d24924b179adf1c2e6f4ea6fb866747e84baea6b2ffaa8bff982ce102",
        image = "index.docker.io/library/docker",
    )

    oci_pull(
        name = "legacy_executor-vm_base",
        digest = "sha256:4b23a8bbfa9e1f5c80b167e59c7f0d07e40b4af52494c22da088a1c97925a3e2",
        image = "index.docker.io/sourcegraph/executor-vm",
    )

    # Please review the changes in /usr/local/share/postgresql/postgresql.conf.sample
    # If there is any change, you should ping @release-team
    # who will make sure changes are reflected in our deploy repository
    oci_pull(
        name = "legacy_postgres-12-alpine_base",
        # IMPORTANT: Only update to Postgres 12.X Alpine linux/x86_64 images, and update the tag below
        # (Bazel doesn't allow both tags and hashes)
        # docker pull --platform linux/x86_64 postgres:12.18-alpine3.18
        digest = "sha256:a7b33f6dc44abdd049d666ee8d919c54642a0b36563c28cf0849b744997da0a8",
        image = "index.docker.io/library/postgres",
        platforms = ["linux/amd64"],
    )
