load("@io_bazel_rules_go//go:def.bzl", "go_cross_binary")
load("@rules_pkg//:pkg.bzl", "pkg_tar")

def get_last_segment(path):
    segments = path.split("/")
    last_segment = segments[-1]
    return last_segment

def container_dependencies(targets):
    for target in targets:
        name = get_last_segment(target)
        platformed_name = "{}_linux_amd64".format(name)

        go_cross_binary(
            name = platformed_name,
            platform = "@zig_sdk//platform:linux_amd64",
            target = target,
        )

        pkg_tar(
            name = "tar_{}".format(platformed_name),
            srcs = [":{}".format(platformed_name)],
            remap_paths = { "/{}".format(platformed_name): "/usr/local/bin/{}".format(name) }
        )

def dependencies_tars(targets):
    tars = []
    for target in targets:
        name = get_last_segment(target)
        tars.append(":tar_{}_linux_amd64".format(name))

    return tars

