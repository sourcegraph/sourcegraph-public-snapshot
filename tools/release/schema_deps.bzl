"""
This module defines the third party dependency containing all database schemas that the
migrator use to handle migrations. See the README.md in this folder for reference.
"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def schema_deps():
    http_file(
        name = "schemas_archive",
        urls = ["https://storage.googleapis.com/schemas-migrations/dist/schemas-v5.2.1.tar.gz"],
        sha256 = "3ec54f2d132ba5fc4f084f3bc76650f1c759ab32b5b73aba2ac9df91098ffeaf",
    )
