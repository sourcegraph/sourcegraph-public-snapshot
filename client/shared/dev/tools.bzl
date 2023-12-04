"Bazel macros for building frontend packages"

load("@aspect_rules_js//js:defs.bzl", "js_library")
load("@npm//:typed-scss-modules/package_json.bzl", types_scss_modules_bin = "bin")

def module_style_typings(name, deps = []):
    """Generate TypeScript types for module.scss files.

    Automatically globs all module.scss files in the Bazel package and
    outputs equivalent .d.ts files next to the styles in the output
    tree.

    Args:
        name: Name of the target
        deps: Additional dependencies imported by any of the source style files.
           These could be other .scss files or node_modules deps.
    """

    srcs = native.glob(["src/**/*.module.scss"])
    outs = ["%s.d.ts" % src for src in srcs]

    types_scss_modules_bin.tsm(
        name = "%s_types" % name,
        srcs = srcs + deps,
        outs = outs,
        args = [
            "--logLevel",
            "error",
            "%s/**/*.module.scss" % native.package_name(),
            "--includePaths",
            "client",
            "node_modules",
        ],
    )

    js_library(
        name = name,
        srcs = [":%s_types" % name],
    )
