"""This module contains definitions for dealing with stubs in the source tree.
"""

load("@aspect_bazel_lib//lib:write_source_files.bzl", "write_source_files")
load("@aspect_bazel_lib//lib:copy_to_directory.bzl", "copy_to_directory")
load("@aspect_bazel_lib//lib:directory_path.bzl", "make_directory_path")

def write_proto_stubs_to_source(name, target, output_files):
    native.filegroup(
        name = name,
        srcs = [target],
        output_group = "go_generated_srcs",
    )

    copy_to_directory(
        name = name + "_flattened",
        srcs = [name],
        root_paths = ["**"],
    )

    write_source_files(
        name = "write_" + name,
        files = {
            output_file: make_directory_path(output_file + "_directory_path", name + "_flattened", output_file)
            for output_file in output_files
        },
    )
