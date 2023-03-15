"Schema generation rule"

load("@aspect_rules_js//js:defs.bzl", "js_run_binary")
load("@bazel_skylib//lib:collections.bzl", "collections")

def generate_schema(name, out):
    """Generate TypeScript types for a JSON schema with target name //schema:<name>.

    Args:
        name: name of the schema file under schemas/ minus the extension
        out: output schema file
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
        outs = [out],
        args = [
            name,
        ],
        tool = "//client/shared/dev:generate_schema",
    )
