load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def schema_deps():
    http_file(
        name = "schemas_archive",
        urls = ["https://storage.googleapis.com/schemas-migrations/dist/schemas-v5.2.123456.tar.gz"],
        sha256 = "b1736a978037c584df1ebc1e16e0e72413eb943df180c46020011c8a71d7e74e",
    )
