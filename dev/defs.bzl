load("@bazel_skylib//lib:partial.bzl", "partial")
load("@bazel_skylib//rules:expand_template.bzl", "expand_template")
load("@aspect_rules_jest//jest:defs.bzl", _jest_test = "jest_test")
load("@aspect_rules_js//npm:defs.bzl", _npm_package = "npm_package")
load("@aspect_rules_ts//ts:defs.bzl", _ts_project = "ts_project")
load(":sass.bzl", _sass = "sass")
load(":babel.bzl", _babel = "babel")

sass = _sass

def ts_project(name, deps = [], **kwargs):
    deps = deps + [
        "//:node_modules/tslib",
    ]

    visibility = kwargs.pop("visibility", ["//visibility:public"])

    common_kwargs = {
        "tags": kwargs.get("tags", []),
        "visibility": visibility,
        "testonly": kwargs.get("testonly", None),
    }

    # Add standard test libraries for the repo test frameworks
    if common_kwargs.get("testonly"):
        deps = deps + [d for d in [
            "//:node_modules/@types/jest",
            "//:node_modules/@types/mocha",
            "//:node_modules/@types/testing-library__jest-dom",
        ] if not d in deps]

    tsconfig = kwargs.pop("tsconfig", "//:tsconfig")

    # Default arguments for ts_project.
    _ts_project(
        name = name,
        deps = deps,

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
            tsconfig = tsconfig,
            **common_kwargs
        ),
        supports_workers = False,

        # Allow any other args
        **kwargs
    )

def npm_package(name, srcs = [], **kwargs):
    replace_prefixes = kwargs.pop("replace_prefixes", {})

    # Modifications to package.json
    # TODO(bazel): remove when package.json can be updated in source
    srcs_no_pkg = [s for s in srcs if s != "package.json"]
    if len(srcs) > len(srcs_no_pkg):
        expand_template(
            name = "_updated-package-json",
            template = "package.json",
            out = "_updated-package.json",
            substitutions = {
                # TODO(bazel): remove use of .ts in package.json files
                "src/index.ts": "src/index.js",
            },
        )
        replace_prefixes["_updated-package.json"] = "package.json"
        srcs = srcs_no_pkg + ["_updated-package.json"]

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
        data = data + native.glob(["**/__snapshots__/**/*.snap"]),
        snapshots = True,
        **kwargs
    )
