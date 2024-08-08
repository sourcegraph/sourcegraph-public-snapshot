"""
Various helpers to help with server image building
"""

load("//dev:oci_defs.bzl", "pkg_tar")

def get_last_segment(path):
    """
    returns part of a bazel path - it will return binary in //something:binary

    Args:
        path: the path from which the last segment should be extract from
    Returns:
        Returns the last part found after `:`. If no `:` is found then the last part after the last '/' is returned
    """
    segments = path.split("/")
    last_segment = segments[-1]

    s = last_segment.split(":")
    if len(s) == 1:
        return last_segment
    else:
        return s[-1]

def container_dependencies(targets):
    """
    creates pkg_tar rules for all given targets

    for all the given targets this will create a pkg_tar rule named 'tar_<name>` where
    the target is added as well as the path of the target output is remapped to be at /usr/local/bin

    Args:
        targets: list of targets for which pkg_tar rules should be generated for
    """
    for target in targets:
        name = get_last_segment(target)

        pkg_tar(
            name = "tar_{}".format(name),
            srcs = [target],
            remap_paths = {"/{}".format(name): "/usr/local/bin/{}".format(name)},
        )

def dependencies_tars(targets):
    """
    for all the given targets it returns a list of the corresponding `:tar_<name>` targets

    Args:
        targets: list of targets
    Returns:
        list of ':tar_<name>' targets
    """
    tars = []
    for target in targets:
        name = get_last_segment(target)
        tars.append(":tar_{}".format(name))

    return tars
