"Schema generation rule"

load("@aspect_rules_js//js:defs.bzl", "js_run_binary")

def build_code_intel_extensions(name, out):
    """ Download code-intel extension bundles from GitHub.

    Args:
        name: target name
        out: output revisions folder
    """
    js_run_binary(
        name = name,
        chdir = native.package_name(),
        out_dirs = [out],
        log_level = "info",
        silent_on_success = False,
        args = [
            "$(execpath @sourcegraph_extensions_bundle//:bundle)",
            out,
        ],
        srcs = ["@sourcegraph_extensions_bundle//:bundle"],
        copy_srcs_to_bin = False,
        tool = "//client/shared/dev:build_code_intel_extensions",
    )
