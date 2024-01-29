load("@aspect_bazel_lib//lib:yq.bzl", "yq")


def _go_mockgen(ctx):


go_mockgen = rule(
    implementation = _go_mockgen,
)

def _go_mockgen_config(rctx):

    for config in rctx.attr.configs:
        rctx.read(rctx.path(config))

go_mockgen_config = repository_rule(
    implementation = _go_mockgen_config,
)
