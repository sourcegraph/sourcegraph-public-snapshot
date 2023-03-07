"Bazel rules"

load("@bazel_skylib//lib:partial.bzl", "partial")
load("@bazel_skylib//rules:expand_template.bzl", "expand_template")
load("@aspect_rules_js//npm:defs.bzl", _npm_package = "npm_package")
load("@aspect_rules_ts//ts:defs.bzl", _ts_project = "ts_project")
load("@aspect_rules_jest//jest:defs.bzl", _jest_test = "jest_test")
load(":sass.bzl", _sass = "sass")
load(":babel.bzl", _babel = "babel")

# Experiment using raw babel instead of tsc.
# TODO(bazel): ts_project(tsc=babel) currently still runs tsc to carry
# .d.ts when going into an npm_package().
# For now use raw babel without any type-checking
USE_RAW_BABEL = True

sass = _sass

def ts_project(name, deps = [], **kwargs):
    """A wrapper around ts_project

    Args:
        name: A unique name for this target

        deps: A list of dependencies

        **kwargs: Additional arguments to pass to ts_project
    """
    deps = deps + [
        "//:node_modules/tslib",
    ]

    visibility = kwargs.pop("visibility", ["//visibility:public"])

    # Add standard test libraries for the repo test frameworks
    if kwargs.get("testonly", False):
        deps = deps + [d for d in [
            "//:node_modules/@types/jest",
            "//:node_modules/@types/mocha",
            "//:node_modules/@types/testing-library__jest-dom",
        ] if not d in deps]

    if USE_RAW_BABEL:
        _ts_project_babel_impl(name, deps = deps, visibility = visibility, **kwargs)
    else:
        _ts_project_tsc_impl(name, deps = deps, visibility = visibility, **kwargs)

# ts_project() as an alias to babel with no type-checking or dts
def _ts_project_babel_impl(name, **kwargs):
    # no tsconfig with babel
    kwargs.pop("tsconfig", None)
    kwargs.pop("allow_js", None)
    kwargs.pop("incremental", None)
    kwargs.pop("composite", None)
    kwargs.pop("declaration_map", None)
    kwargs.pop("emit_declaration_only", None)

    _babel(
        name = name,
        **kwargs
    )

# ts_project() with full type-checking etc
def _ts_project_tsc_impl(name, visibility, **kwargs):
    tsconfig = kwargs.pop("tsconfig", "//:tsconfig")

    # Default arguments for ts_project.
    _ts_project(
        name = name,
        deps = kwargs.pop("deps", []),

        # tsconfig options, default to the root
        tsconfig = tsconfig,
        composite = kwargs.pop("composite", True),
        declaration = kwargs.pop("declaration", True),
        declaration_map = kwargs.pop("declaration_map", True),
        resolve_json_module = kwargs.pop("resolve_json_module", True),
        source_map = kwargs.pop("source_map", True),
        preserve_jsx = kwargs.pop("preserve_jsx", None),
        visibility = visibility,

        # use babel as the transpiler
        transpiler = partial.make(
            _babel,
            module = kwargs.pop("module", None),
            tags = kwargs.get("tags", []),
            visibility = visibility,
            testonly = kwargs.get("testonly", None),
        ),
        supports_workers = False,

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

    package_type = kwargs.pop("type", "module")

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
                "\"name\"": "\"type\": \"%s\",\n  \"name\"" % package_type,
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

def jest_test(name, data = [], **kwargs):
    _jest_test(
        name = name,
        config = "//:jest_config",
        snapshots = kwargs.pop("snapshots", True),
        data = data + native.glob(["**/__fixtures__/**/*"]),
        **kwargs
    )
