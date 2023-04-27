load("@rules_oci//oci:pull.bzl", "oci_pull")

def oci_deps():
    oci_pull(
        name = "wolfi_base",
        digest = "sha256:bb939c611ced27e5e566ad2a402a9f030fca949bbd351a8f84fcb68f4e790e5d",
        image = "europe-central2-docker.pkg.dev/sourcegraph-security-logging/public-wolfi-test/wolfi-sourcegraph-dev-base",
        # platforms = [
        #     "linux/amd64",
        #     "linux/arm64",
        # ],
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:8f9f940b3173023c5aeea756c61b03a12ec111905316df09ea7eb7ef4ed81570",
        image = "europe-central2-docker.pkg.dev/sourcegraph-security-logging/public-wolfi-test/wolfi-symbols-base",
    )
