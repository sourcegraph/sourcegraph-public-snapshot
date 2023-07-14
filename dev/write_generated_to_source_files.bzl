load("@aspect_bazel_lib//lib:directory_path.bzl", "make_directory_path")
load("@aspect_bazel_lib//lib:copy_to_directory.bzl", "copy_to_directory")
load("@aspect_bazel_lib//lib:write_source_files.bzl", "write_source_files")
load("@bazel_skylib//lib:paths.bzl", "paths")

def write_generated_to_source_files(name, src, files, dest = "", strip_prefix = "", verbose_copy=True, **kwargs):

    currentTarget = native.package_relative_label(name)
    srcTarget = Label(src)
    copy_opts = {
        "name" : "copy_" + name,
        "srcs" : [src],
    }

    # when we're in the same package as the src target, we don't need to change the root path
    # BUT when we're in different packages we need to replace the src target package root, since
    # all it's outputs path will have the src package as part of the path. So set root_paths to
    # the src target package path, which means it will be removed as part of copying
    if currentTarget.package != srcTarget.package:
        copy_opts["root_paths"] = [
            srcTarget.package
        ]
        # copy_opts["replace_prefixes"] = {
        #     srcTarget.package: "",
        # }

    # We use a copy_to_directory macro so write_source_files inputs and outputs are not at the same
    # path, which enables the write_doc_files_diff_test to work.
    copy_to_directory(verbose=verbose_copy,**copy_opts)

    def dest_path(value, strip_prefix):
        target = value.removeprefix(strip_prefix)
        if dest:
            target = paths.join(dest, target)

        print(target)
        return target

    write_source_files(
        name = name,
        files =  {
            dest_path(out, strip_prefix): make_directory_path(
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
