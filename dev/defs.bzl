load("@bazel_skylib//rules:build_test.bzl", "build_test")
load("@aspect_rules_jest//jest:defs.bzl", _jest_test = "jest_test")
load("@aspect_rules_js//npm:defs.bzl", _npm_package = "npm_package")
load("@aspect_rules_ts//ts:defs.bzl", _ts_config = "ts_config", _ts_project = "ts_project")
load("@npm//:sass/package_json.bzl", sass_bin = "bin")
load("@bazel_skylib//rules:expand_template.bzl", "expand_template")

def ts_project(name, deps = [], **kwargs):
    deps = deps + [
        "//:node_modules/tslib",
    ]

    testonly = kwargs.pop("testonly", False)

    # Add standard test libraries for the repo test frameworks
    if testonly:
        deps = deps + [d for d in [
            "//:node_modules/@types/jest",
            "//:node_modules/@types/mocha",
            "//:node_modules/@types/testing-library__jest-dom",
        ] if not d in deps]

    # Default arguments for ts_project.
    _ts_project(
        name = name,
        deps = deps,

        # tsconfig options, default to the root
        tsconfig = kwargs.pop("tsconfig", "//:tsconfig"),
        composite = kwargs.pop("composite", True),
        declaration = kwargs.pop("declaration", True),
        declaration_map = kwargs.pop("declaration_map", True),
        resolve_json_module = kwargs.pop("resolve_json_module", True),
        source_map = kwargs.pop("source_map", True),

        # Rule options
        visibility = kwargs.pop("visibility", ["//visibility:public"]),
        testonly = testonly,
        supports_workers = False,

        # Allow any other args
        **kwargs
    )

    build_test(
        name = "%s_build_test" % name,
        targets = [name],
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

def sass(name, srcs, deps = [], **kwargs):
    sass_bin.sass(
        name = name,
        srcs = srcs + deps,
        outs = [src.replace(".scss", ".css") for src in srcs],
        args = [
            "--load-path=client",
            "--load-path=node_modules",
        ] + [
            "$(execpath {}):{}/{}".format(src, native.package_name(), src.replace(".scss", ".css"))
            for src in srcs
        ],

        # Default visiblity
        visibility = kwargs.pop("visibility", ["//visibility:public"]),

        # Allow any other args
        **kwargs
    )

def jest_test(name, data = [], **kwargs):
    # TODO(bazel): the config param must be a single file. Must manually
    # declare config dependencies and pass as 'data'.
    config = "//:jest.config.base"
    data = data + ["//:jest_config"]

    _jest_test(
        name = name,
        config = config,
        data = data,
        **kwargs
    )
