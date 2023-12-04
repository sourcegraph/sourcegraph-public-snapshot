"Sass compilation rules"

load("@npm//client/build-config:sass/package_json.bzl", sass_bin = "bin")
load("@npm//client/build-config:postcss-cli/package_json.bzl", postcss_bin = "bin")

# A filename for the intermediate file between sass + postcss
def _sass_out(n):
    return n.replace(".scss", "_scss.css")

# SASS and PostCSS
def sass(name, srcs, deps = [], runtime_deps = [], **kwargs):
    """Runs SASS and PostCSS on sass inputs

    Args:
        name: A unique name for the terminal target

        srcs: A list of .scss sources

        deps: A list of dependencies

        runtime_deps: A list of runtime_dependencies

        **kwargs: Additional arguments
    """
    visibility = kwargs.pop("visibility", None)

    sass_bin.sass(
        name = "_%s_sass" % name,
        srcs = srcs + deps,
        outs = [_sass_out(src) for src in srcs] + ["%s.map" % _sass_out(src) for src in srcs],
        args = [
            "--load-path=client",
            "--load-path=node_modules",
        ] + [
            "$(execpath {}):{}/{}".format(src, native.package_name(), _sass_out(src))
            for src in srcs
        ],
        visibility = ["//visibility:private"],
    )

    for src in srcs:
        _postcss(
            name = src.replace(".scss", "_css"),
            src = _sass_out(src),
            out = src.replace(".scss", ".css"),

            # Same visibility as filegroup of outputs
            visibility = visibility,
        )

    native.filegroup(
        name = name,
        srcs = [src.replace(".scss", ".css") for src in srcs] + [src.replace(".scss", ".css.map") for src in srcs],
        visibility = visibility,
        data = runtime_deps,

        # Allow any other args
        **kwargs
    )

def _postcss(name, src, out, **kwargs):
    postcss_bin.postcss(
        name = name,
        srcs = [
            src,
            "%s.map" % src,
            "//:postcss_config_js",
        ],
        outs = [out, out + ".map"],
        # rules_js runs in the execroot under the output tree in bazel-out/[arch]/bin
        args = [
            "../../../$(execpath %s)" % src,
            "--config",
            "../../../$(execpath //:postcss_config_js)",
            "--map",
            "--output",
            "../../../$(@D)/%s" % out,
        ],
        **kwargs
    )
