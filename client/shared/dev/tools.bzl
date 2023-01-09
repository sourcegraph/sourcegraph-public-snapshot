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

    srcs = native.glob([
        "src/**/*.module.scss",
        "webview/**/*.module.scss",  # JetBrains client folder structure slightly different
    ])

    outs = ["%s.d.ts" % src for src in srcs]

    js_library(
        name = "%s_sources" % name,
        srcs = srcs,
    )

    types_scss_modules_bin.tsm(
        name = name,
        srcs = [":%s_sources" % name] + deps,
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
