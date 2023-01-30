load("@bazel_skylib//lib:paths.bzl", "paths")
load("@npm//:@babel/cli/package_json.bzl", "bin")

def babel(name, srcs, tsconfig, **kwargs):
    # rules_js runs in the execroot under the output tree in bazel-out/[arch]/bin
    execroot = "../../.."

    ts_srcs = []
    outs = []
    data = []

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
        map_out = js_out + ".map"

        outs.append(js_out)
        outs.append(map_out)

    # see https://babeljs.io/docs/en/babel-cli
    args = [
        native.package_name(),
        "--source-maps",
        "--config-file",
        "{}/$(location {})".format(execroot, "//:babel_config"),
        "--presets=@babel/preset-typescript",
        "--extensions",
        ".ts,.tsx",
        "--out-dir",
        "{}/{}".format(".", native.package_name()),
    ]

    bin.babel(
        name = "{}_lib".format(name),
        progress_message = "Compiling {}".format(native.package_name()),
        srcs = ts_srcs + [
            "//:babel_config",
            "//:package_json",
            tsconfig,
        ],
        outs = outs,
        args = args,
        env = {
            "TSCONFIG": "{}/$(location {})".format(execroot, tsconfig),
        },
        **kwargs
    )

    # The target whose default outputs are the js files which ts_project() will reference
    native.filegroup(
        name = name,
        srcs = data + outs,
    )
