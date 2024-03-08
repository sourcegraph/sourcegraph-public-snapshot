"""
This module defines the third party dependency containing all database schemas that the
migrator use to handle migrations. See the README.md in this folder for reference.
"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def schema_deps():
    http_file(
        name = "schemas_archive",
        urls = ["https://storage.googleapis.com/schemas-migrations/dist/schemas-v5.3.2.tar.gz"],
        sha256 = "f3d23671f8d4aff572ff930c856f644c3fc5c507883dd5f11db95ee205712384",
    )
