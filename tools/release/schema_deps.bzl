"""
This module defines the third party dependency containing all database schemas that the
migrator use to handle migrations. See the README.md in this folder for reference.
"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def schema_deps():
    http_file(
        name = "schemas_archive",
        urls = ["https://storage.googleapis.com/schemas-migrations/dist/schemas-v5.2.2.tar.gz"],
        sha256 = "a60fb3311e164eb4b3061e56f5049c4b6324ccb1a301065b1c773b3dd04d2334",
    )
