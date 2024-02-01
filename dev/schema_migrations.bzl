def _schema_migrations(rctx):
    gsutil_path = rctx.path(
        Label("@gcloud-{}-{}//:gsutil".format({
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

    rctx.file("BUILD.bazel", content = """
package(default_visibility = ["//visibility:public"])

exports_files(["archives"])

filegroup(
    name = "srcs",
    srcs = glob(["**"]),
)
""")

    rctx.execute(["mkdir", "archives"])
    rctx.report_progress("Downloading schema migrations from GCS")
    result = rctx.execute([
        gsutil_path,
        "-m",
        "cp",
        "gs://schemas-migrations/migrations/*",
        "archives",
    ], timeout = 60, environment = {
        "CLOUDSDK_CORE_PROJECT": "sourcegraph-ci",
    })
    if result.return_code != 0:
        fail("Failed to download schema migrations from GCS: {}".format(result.stderr))

schema_migrations = repository_rule(
    implementation = _schema_migrations,
)
