"""
This module defines the third party dependency containing all database schemas that the
migrator use to handle migrations. See the README.md in this folder for reference.
"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def schema_deps():
    http_file(
        name = "schemas_archive",
        urls = ["https://storage.googleapis.com/schemas-migrations/dist/schemas-v5.3.0.tar.gz"],
        sha256 = "6c1c855b7636fc60e2f08f0961e05c019df7ef431a766f485855f871c7c1122f",
    )
