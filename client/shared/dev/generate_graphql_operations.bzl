"GraphQL generation rule"

load("@aspect_rules_js//js:defs.bzl", "js_run_binary")

def generate_graphql_operations(name, interface_name, srcs, out, **kwargs):
    """Generate a graphql operations Typescript interface.

    Args:
        name: Name of the target
        interface_name: Name of the generated TypeScript interface. Must match one of the
            interfaces precoded in generateGraphQlOperations.js.
        srcs: Files required by the glob to generate the interface. Should include all files
            in the glob pattern specified in generateGraphQlOperations.js. The files should be
            produced in or copied to the Bazel output tree, so it's recommended to use js_library
            with a glob pattern equivalent to the one in generateGraphQlOperations.js.
        out: Name of the typescript file to output.
        **kwargs: general args
    """
    js_run_binary(
        name = name,
        outs = [out],
        srcs = srcs,
        args = [
            interface_name,
            out,
        ],
        chdir = native.package_name(),
        tool = "//client/shared/dev:generate_graphql_operations",
        **kwargs
    )
