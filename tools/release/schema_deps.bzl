"""
This module defines the third party dependency containing all database schemas that the
migrator use to handle migrations. See the README.md in this folder for reference.
"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def schema_deps():
    http_file(
        name = "schemas_archive",
        urls = ["https://storage.googleapis.com/schemas-migrations/dist/schemas-v5.2.7.tar.gz"],
        sha256 = "f54994171850e3475a9b446c29e4db4c1ea0ea039ce1cc5355f71f5b4725b117",
    )
