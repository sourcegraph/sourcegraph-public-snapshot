"Bazel rules"

load("@aspect_rules_swc//swc:defs.bzl", "swc")
load("@bazel_skylib//lib:partial.bzl", "partial")
load("@bazel_skylib//rules:expand_template.bzl", "expand_template")
load("@aspect_rules_js//js:defs.bzl", "js_library")
load("@aspect_rules_js//npm:defs.bzl", _npm_package = "npm_package")
load("@aspect_rules_ts//ts:defs.bzl", _ts_project = "ts_project")
load("@aspect_rules_js//js:defs.bzl", "js_binary")
load("//dev:eslint.bzl", "eslint_test_with_types", "get_client_package_path")
load("@npm//:vitest/package_json.bzl", vitest_bin = "bin")
load(":sass.bzl", _sass = "sass")

sass = _sass

# TODO move this to `ts_project.bzl`
def ts_project(name, srcs = [], deps = [], module = "es6", **kwargs):
    """A wrapper around ts_project

    Args:
        name: A unique name for this target

        srcs: A list of source files

        deps: A list of dependencies

        module: The module type to use for the project (es6 or commonjs)

        **kwargs: Additional arguments to pass to ts_project
    """

    # Add the ESLint test target which lints all srcs of the `ts_project`.
    eslint_test_with_types(
        name = "%s_eslint" % name,
        srcs = srcs,
        deps = deps,
        config = "//{}:eslint_config".format(get_client_package_path()),
    )

    deps = deps + [
        "//:node_modules/tslib",
    ]

    visibility = kwargs.pop("visibility", ["//visibility:public"])

    # Add standard test libraries for the repo test frameworks
    testonly = kwargs.get("testonly", False)
    if testonly:
        deps = deps + [d for d in [
            "//:node_modules/@types/mocha",
            "//:node_modules/vitest",
        ] if not d in deps]

    transpiler = partial.make(
        swc,
        swcrc = kwargs.pop("swcrc", "//:.swcrc"),
        # Test code using jest.mock needs to be transpiled to CommonJS.
        args = ["--config-json", '{"module": {"type": "commonjs"}}'] if module == "commonjs" else [],
    )

    # Default arguments for ts_project.
    _ts_project(
        name = name,
        srcs = srcs,
        deps = deps,
        transpiler = transpiler,

        # tsconfig options, default to the root
        tsconfig = kwargs.pop("tsconfig", "//:tsconfig"),
        composite = kwargs.pop("composite", True),
        declaration = kwargs.pop("declaration", True),
        declaration_map = kwargs.pop("declaration_map", True),
        resolve_json_module = kwargs.pop("resolve_json_module", True),
        source_map = kwargs.pop("source_map", True),
        preserve_jsx = kwargs.pop("preserve_jsx", None),
        visibility = visibility,
        supports_workers = 0,

        # Allow any other args
        **kwargs
    )

def npm_package(name, srcs = [], **kwargs):
    """A wrapper around npm_package

    Args:
        name: A unique name for this target

        srcs: A list of sources

        **kwargs: Additional arguments to pass to npm_package
    """
    replace_prefixes = kwargs.pop("replace_prefixes", {})

    package_type = kwargs.pop("type", "commonjs")

    # Modifications to package.json
    # TODO(bazel): remove when package.json can be updated in source
    srcs_no_pkg = [s for s in srcs if s != "package.json"]
    if len(srcs) > len(srcs_no_pkg):
        updated_pkg = "_%s_package" % name
        updated_pkg_json = "%s.json" % updated_pkg

        # remove references to .ts in package.json files
        expand_template(
            name = updated_pkg,
            template = "package.json",
            out = updated_pkg_json,
            substitutions = {
                "src/index.ts": "src/index.js",
                # "\"name\"": "\"type\": \"%s\",\n  \"name\"" % package_type,
            },
        )

        replace_prefixes[updated_pkg_json] = "package.json"
        srcs = srcs_no_pkg + [updated_pkg_json]

    _npm_package(
        name = name,
        srcs = srcs,
        replace_prefixes = replace_prefixes,

        # Default visiblity
        visibility = kwargs.pop("visibility", ["//visibility:public"]),

        # Allow any other args
        **kwargs
    )

def vitest_test(name, data = [], **kwargs):
    vitest_config = "%s_vitest_config" % name

    js_library(
        name = vitest_config,
        testonly = True,
        srcs = ["vitest.config.ts"],
        deps = ["//:vitest_config", "//:node_modules/vitest"],
        data = data,
    )

    vitest_bin.vitest_test(
        name = name,
        args = [
            "run",
            "--reporter=default",
            "--color",
            "--update=false",
            "--config=$(location :%s)" % vitest_config,
        ],
        data = data + native.glob(["**/__fixtures__/**/*"]) + [
            ":%s" % vitest_config,
            "//:node_modules/happy-dom",
            "//:node_modules/jsdom",
        ],
        env = {"BAZEL": "1", "CI": "1"},
        patch_node_fs = False,
        tags = kwargs.pop("tags", []) + ["no-sandbox"],
        timeout = kwargs.pop("timeout", "short"),
        **kwargs
    )

def ts_binary(name, entry_point, data = [], env = {}, **kwargs):
    """A wrapper around js_binary that invokes a TypeScript entrypoint using ts-node."""
    js_binary(
        name = name,
        entry_point = entry_point,
        data = data + ["//:node_modules/ts-node"],
        env = dict(env, **{"TS_NODE_TRANSPILE_ONLY": "1"}),
        node_options = ["--require", "ts-node/register"],
        **kwargs
    )
