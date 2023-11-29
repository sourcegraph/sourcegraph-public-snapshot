"""
This module defines the third party dependency containing all database schemas that the
migrator use to handle migrations. See the README.md in this folder for reference.
"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def schema_deps():
    http_file(
        name = "schemas_archive",
        urls = ["https://storage.googleapis.com/schemas-migrations/dist/schemas-v5.2.3.tar.gz"],
        sha256 = "c5aec72d528c0b3803070ccba58049c42f9b2618c9dba367dffe106d30f8f6fe",
    )
