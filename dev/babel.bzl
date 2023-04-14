"Babel rule"

load("@aspect_rules_js//js:defs.bzl", "js_library")
load("@npm//:@babel/cli/package_json.bzl", "bin")

def babel(name, srcs, module = None, use_preset_env = True, **kwargs):
    """A wrapper around Babel CLI

    Args:
        name: A unique name for this target

        srcs: A list of sources

        module: If specified, sets BABEL_MODULE environment variable to this value

        use_preset_env: Controls if we transpile TS sources with babel-preset-env.
        If set to False, sets the DISABLE_PRESET_ENV environment variable to "true".

        **kwargs: Additional arguments to pass to the rule
    """

    # rules_js runs in the execroot under the output tree in bazel-out/[arch]/bin
    execroot = "../../.."

    visibility = kwargs.pop("visibility", ["//visibility:public"])
    source_map = kwargs.pop("source_map", True)

    ts_srcs = []
    outs = []
    data = kwargs.pop("data", [])
    deps = kwargs.pop("deps", [])

    # Collect the srcs to compile and expected outputs
    for src in srcs:
        # JSON does not need to be compiled
        if src.endswith(".json"):
            data.append(src)
            continue

        # dts are only for type-checking and not to be compiled
        if src.endswith(".d.ts"):
            continue

        if not (src.endswith(".ts") or src.endswith(".tsx")):
            fail("babel example transpiler only supports source .ts[x] files, got: %s" % src)

        ts_srcs.append(src)

        # Predict the output paths where babel will write
        js_out = src.replace(".tsx", ".js").replace(".ts", ".js")

        outs.append(js_out)
        if source_map:
            outs.append(js_out + ".map")

    # see https://babeljs.io/docs/en/babel-cli
    args = [
        native.package_name(),
        "--config-file",
        "{}/$(location {})".format(execroot, "//:babel_config"),
        "--source-maps",
        "true" if source_map else "false",
        "--extensions",
        ".ts,.tsx",
        "--out-dir",
        "{}/{}".format(".", native.package_name()),
    ]

    env = {}
    if module != None:
        env["BABEL_MODULE"] = module

    if use_preset_env == False:
        env["DISABLE_PRESET_ENV"] = "true"

    bin.babel(
        name = "{}_lib".format(name),
        progress_message = "Compiling {}:{}".format(native.package_name(), name),
        srcs = ts_srcs + [
            "//:babel_config",
            "//:package_json",
            "//:browserslist",
        ],
        outs = outs,
        args = args,
        env = env,
        **kwargs
    )

    js_library(
        name = name,
        srcs = outs,
        deps = deps,
        data = data,
        visibility = visibility,
    )
