load("@aspect_rules_ts//ts:defs.bzl", _ts_project = "ts_project")
load(":sass.bzl", _sass = "sass")

def ts_project(name, deps = [], **kwargs):
    deps = deps + [
        "//:node_modules/tslib",
    ]

    """Default arguments for ts_project."""
    _ts_project(
        name = name,
        deps = deps,

        # tsconfig options
        tsconfig = "//:tsconfig",
        composite = True,
        declaration = True,
        declaration_map = True,
        resolve_json_module = True,
        source_map = True,

        # Allow any other args
        **kwargs
    )

sass = _sass
