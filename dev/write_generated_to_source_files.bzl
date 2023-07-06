load("@aspect_bazel_lib//lib:directory_path.bzl", "make_directory_path")
load("@aspect_bazel_lib//lib:copy_to_directory.bzl", "copy_to_directory")
load("@aspect_bazel_lib//lib:write_source_files.bzl", "write_source_files")

def write_generated_to_source_files(name, src, files, **kwargs):
    # We use a copy_to_directory macro so write_source_files inputs and outputs are not at the same
    # path, which enables the write_doc_files_diff_test to work.
    copy_to_directory(
        name = "copy_" + name,
        srcs = [src]
    )

    write_source_files(
        name = name,
        files = {
            out: make_directory_path(
                out + "_directory_path",
                "copy_" + name,
                out,
            )
            for out in files
        },
        suggested_update_target = "//dev:write_all_generated",
        visibility = ["//visibility:public"],
        **kwargs,
    )
