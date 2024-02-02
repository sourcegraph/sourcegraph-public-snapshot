load("@aspect_bazel_lib//:lib.bzl", "run_binary")

def go_mockgen(name, out_file, pkg, file_prefix):
    run_binary(
        name = "mockgen-{}".format(name),
        mnemonic = "GoMockgen",
        tool = "@com_github_derision_test_go_mockgen//cmd/go-mockgen:go-mockgen",
        args = ["--filename", out_file, "--package", pkg, "--file-prefix", file_prefix],
        outs = [out_file],
    )

    # TODO: copy to source tree

def _go_mockgen_config(rctx):
    yq_path = rctx.path(
        Label("@yq-{}-{}//file:downloaded".format({
            "mac os x": "darwin",
            "linux": "linux",
        }[rctx.os.name], {
            "aarch64": "arm64",
            "arm64": "arm64",
            "amd64": "amd64",
            "x86_64": "amd64",
            "x86": "amd64",
        }[rctx.os.arch])),
    )

    mockgen_path = rctx.path(Label("//:mockgen.yaml"))

    # rctx.file("BUILD", "[]")
    rctx.template("BUILD", Label("//dev:go-mockgen.tpl"))

    base_config = yaml_to_json(rctx, yq_path, mockgen_path)

    # prefix for the top of each generated file
    file_prefix = base_config["file-prefix"]

    for config in base_config["include-config-paths"]:
        # parse each to determine what to generate and where
        c = yaml_to_json(rctx, yq_path, rctx.path(Label("//:{}".format(config))))

    # TODO: finish the owl

def yaml_to_json(rctx, yq_path, path):
    res = rctx.execute([yq_path, "-p", "yaml", "-o", "json", path])
    if res.return_code != 0:
        fail("failed to run yq: {}".format(res.stderr))

    return json.decode(res.stdout)

go_mockgen_config = repository_rule(
    implementation = _go_mockgen_config,
)
