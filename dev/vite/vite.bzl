"Vite rule"

load("@aspect_bazel_lib//lib:copy_to_bin.bzl", "copy_file_to_bin_action", "copy_files_to_bin_actions")
load("@aspect_bazel_lib//lib:copy_to_directory.bzl", "copy_to_directory")
load("@aspect_bazel_lib//lib:directory_path.bzl", "directory_path")
load("@aspect_rules_js//js:defs.bzl", "js_binary")
load("@aspect_rules_js//js:libs.bzl", "js_lib_helpers")
load("@aspect_rules_js//js:providers.bzl", "JsInfo", "js_info")
load("@bazel_skylib//lib:paths.bzl", "paths")

_ATTRS = {
    "config": attr.label(
        mandatory = True,
        allow_single_file = True,
        doc = """Configuration file used for Vite""",
    ),
    "data": js_lib_helpers.JS_LIBRARY_DATA_ATTR,
    "deps": attr.label_list(
        default = [],
        doc = "A list of direct dependencies that are required to build the bundle",
        providers = [JsInfo],
    ),
    "entry_points": attr.label_list(
        allow_files = True,
        doc = """The bundle's entry points""",
    ),
    "srcs": attr.label_list(
        allow_files = True,
        default = [],
        doc = """Source files to be made available to Vite""",
    ),
    "tsconfig": attr.label(
        mandatory = True,
        allow_single_file = True,
        doc = """TypeScript configuration file used by Vite""",
    ),
    "vite_js_bin": attr.label(
        executable = True,
        doc = "Override the default vite executable",
        cfg = "exec",
    ),
    "env": attr.string_dict(),
}

def _bin_relative_path(ctx, file):
    prefix = ctx.bin_dir.path + "/"
    if file.path.startswith(prefix):
        return file.path[len(prefix):]

    # Since file.path is relative to execroot, go up with ".." starting from
    # ctx.bin_dir until we reach execroot, then join that with the file path.
    up = "/".join([".." for _ in ctx.bin_dir.path.split("/")])
    return up + "/" + file.path

def _output_relative_path(f):
    "Give the path from bazel-out/[arch]/bin to the given File object"
    if f.short_path.startswith("../"):
        return "external/" + f.short_path[3:]
    return f.short_path

def _filter_js(files):
    return [f for f in files if f.extension == "js" or f.extension == "mjs"]

def _vite_project_impl(ctx):
    input_sources = copy_files_to_bin_actions(ctx, ctx.files.srcs)
    entry_points = copy_files_to_bin_actions(ctx, _filter_js(ctx.files.entry_points))
    inputs = entry_points + input_sources + ctx.files.deps

    args = ctx.actions.args()

    output_sources = [getattr(ctx.outputs, o) for o in dir(ctx.outputs)]
    output_sources.append(ctx.actions.declare_directory(ctx.label.name))
    args.add_all(["--outDir", output_sources[0].basename])

    config_file = copy_file_to_bin_action(ctx, ctx.file.config)
    args.add_all(["--config", _output_relative_path(config_file)])
    inputs.append(config_file)

    env = {
        "BAZEL_BINDIR": ctx.bin_dir.path,
    }
    for (key, value) in ctx.attr.env.items():
        env[key] = value

    args.add("build")

    ctx.actions.run(
        executable = ctx.executable.vite_js_bin,
        arguments = [args],
        inputs = depset(
            inputs,
            transitive = [js_lib_helpers.gather_files_from_js_providers(
                targets = ctx.attr.srcs + ctx.attr.deps,
                include_transitive_sources = True,
                include_declarations = False,
                include_npm_linked_packages = True,
            )],
        ),
        outputs = output_sources,
        progress_message = "Building Vite project %s" % (" ".join([_bin_relative_path(ctx, entry_point) for entry_point in entry_points])),
        mnemonic = "Vite",
        env = env,
    )

    npm_linked_packages = js_lib_helpers.gather_npm_linked_packages(
        srcs = ctx.attr.srcs,
        deps = [],
    )

    npm_package_store_deps = js_lib_helpers.gather_npm_package_store_deps(
        # Since we're bundling, only propagate `data` npm packages to the direct dependencies of
        # downstream linked `npm_package` targets instead of the common `data` and `deps` pattern.
        targets = ctx.attr.data,
    )

    output_sources_depset = depset(output_sources)

    runfiles = js_lib_helpers.gather_runfiles(
        ctx = ctx,
        sources = output_sources_depset,
        data = ctx.attr.data,
        # Since we're bundling, we don't propagate any transitive runfiles from dependencies
        deps = [],
    )

    return [
        DefaultInfo(
            files = output_sources_depset,
            runfiles = runfiles,
        ),
        js_info(
            npm_linked_package_files = npm_linked_packages.direct_files,
            npm_linked_packages = npm_linked_packages.direct,
            npm_package_store_deps = npm_package_store_deps,
            sources = output_sources_depset,
            # Since we're bundling, we don't propagate linked npm packages from dependencies since
            # they are bundled and the dependencies are dropped. If a subset of linked npm
            # dependencies are not bundled it is up the the user to re-specify these in `data` if
            # they are runtime dependencies to progagate to binary rules or `srcs` if they are to be
            # propagated to downstream build targets.
            transitive_npm_linked_package_files = npm_linked_packages.direct_files,
            transitive_npm_linked_packages = npm_linked_packages.direct,
            # Since we're bundling, we don't propagate any transitive sources from dependencies
            transitive_sources = output_sources_depset,
        ),
    ]

lib = struct(
    attrs = _ATTRS,
    implementation = _vite_project_impl,
    toolchains = [
        "@rules_nodejs//nodejs:toolchain_type",
    ],
)

_vite_project = rule(
    implementation = _vite_project_impl,
    attrs = _ATTRS,
    toolchains = lib.toolchains,
    doc = """\
Runs Vite in Bazel
""",
)

def vite_project(name, config, data = [], deps = [], entry_points = [], srcs = [], env = {}, **kwargs):
    """Runs Vite in Bazel.

    Args:

      name: A unique name for this rule.

      config: The Vite config file.

      data: Runtime dependencies that are passed to the Vite build.

      deps: Other npm dependencies that are passed to the Vite build.

      entry_points: The entry points to build.

      srcs: Other sources that are passed to the Vite build.

      env: Environment variables to pass to the Vite build.

      **kwargs: Additional arguments
    """
    vite_js_entry_point = "_{}_vite_js_entry_point".format(name)
    node_modules = "//:node_modules"
    directory_path(
        name = vite_js_entry_point,
        directory = "{}/vite/dir".format(node_modules),
        path = "bin/vite.js",
    )

    vite_js_bin = "_{}_vite_js_bin".format(name)
    js_binary(
        name = vite_js_bin,
        data = ["%s/vite" % node_modules],
        entry_point = vite_js_entry_point,
    )

    _vite_project(
        name = name,
        config = config,
        data = data,
        deps = deps,
        entry_points = entry_points,
        srcs = srcs,
        vite_js_bin = vite_js_bin,
        env = env,
        **kwargs
    )

def vite_web_app(name, **kwargs):
    bundle_name = "%s_bundle" % name

    vite_project(name = bundle_name, **kwargs)

    copy_to_directory(
        name = name,
        # flatten static assets
        # https://docs.aspect.build/rules/aspect_bazel_lib/docs/copy_to_directory/#root_paths
        root_paths = ["client/web/dist", "client/web/%s" % bundle_name],
        srcs = ["//client/web/dist/img:img", ":%s" % bundle_name],
        visibility = ["//visibility:public"],
    )
