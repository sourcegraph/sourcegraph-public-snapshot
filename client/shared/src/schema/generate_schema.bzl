load("@aspect_rules_js//js:defs.bzl", "js_run_binary")
load("@bazel_skylib//lib:collections.bzl", "collections")

def generate_schema(name, tool):
    """Generate TypeScript types for a schema.

    Converts the schema with target name //schema:<name> and outputs
    <name>.schema.d.ts.

    Args:
        name: name of the schema file under schemas/ minus the extension
        tool: js_binary that performs the schema generation
    """
    js_run_binary(
        name = name,
        chdir = native.package_name(),
        srcs = collections.uniq(
            [
                "//schema:%s" % name,
                "//schema:json-schema-draft-07",
            ],
        ),
        outs = ["%s.schema.d.ts" % name],
        args = [
            name,
        ],
        tool = tool,
    )
