load("@rules_pkg//:pkg.bzl", "pkg_tar")
load("@rules_pkg//pkg:mappings.bzl", "pkg_attributes", "pkg_filegroup", "pkg_files", "pkg_mkdirs")

def get_last_segment(path):
    segments = path.split("/")
    last_segment = segments[-1]

    s = last_segment.split(":")
    if len(s) == 1:
        return last_segment
    else:
        return s[-1]

def container_dependencies(targets):
    for target in targets:
        name = get_last_segment(target)

        pkg_files(
            name = "files_{}".format(name),
            srcs = [target],
            attributes = pkg_attributes(
                mode = "0555",
            ),
            prefix = "/usr/local/bin",
        )

        pkg_tar(
            name = "tar_{}".format(name),
            srcs = [":files_{}".format(name)],
            # remap_paths = {"/{}".format(name): "/usr/local/bin/{}".format(name)},
            # modes = {"/usr/local/bin/{}".format(name): "0555"},
        )

def dependencies_tars(targets):
    tars = []
    for target in targets:
        name = get_last_segment(target)
        tars.append(":tar_{}".format(name))

    return tars
