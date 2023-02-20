load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file", "http_archive")

def tools_dependencies():
    http_file(
      name = "kubectl_darwin_arm64",
      urls = ["https://dl.k8s.io/release/v1.26.0/bin/darwin/arm64/kubectl"],
      sha256 = "cc7542dfe67df1982ea457cc6e15c171e7ff604a93b41796a4f3fa66bd151f76",
      executable = True,
    )
    http_file(
      name = "kubectl_darwin_amd64",
      urls = ["https://dl.k8s.io/release/v1.26.0/bin/darwin/amd64/kubectl"],
      sha256 = "be9dc0782a7b257d9cfd66b76f91081e80f57742f61e12cd29068b213ee48abc",
      executable = True,
    )
    # TODO linux


    http_file(
      name = "shfmt_darwin_amd64",
      urls = ["https://github.com/mvdan/sh/releases/download/v3.6.0/shfmt_v3.6.0_darwin_amd64"],
      sha256 = "b8c9c025b498e2816b62f0b717f6032e9ab49e725a45b8205f52f66318f17185",
      executable = True,
    )
    http_file(
      name = "shfmt_darwin_arm64",
      urls = ["https://github.com/mvdan/sh/releases/download/v3.6.0/shfmt_v3.6.0_darwin_arm64"],
      sha256 = "633f242246ee0a866c5f5df25cbf61b6af0d5e143555aca32950059cf13d91e0",
      executable = True,
    )

    packer_build_file_content = """
filegroup(
    name = "packer_binary",
    srcs = ["packer"],
    visibility = ["//visibility:public"],
)
    """

    http_archive(
      name = "packer_darwin_arm64",
      urls = ["https://releases.hashicorp.com/packer/1.8.6/packer_1.8.6_darwin_arm64.zip"],
      sha256 = "d7497f1ba3c6221feb4a38a960626afbc4eb26552fc40d941bff57cc2b83bd59",
      build_file_content = packer_build_file_content,
    )

    http_archive(
      name = "packer_darwin_amd64",
      urls = ["https://releases.hashicorp.com/packer/1.8.6/packer_1.8.6_darwin_amd64.zip"],
      sha256 = "0647038561ac14e87bae81e578d575755598346ed3e272ac98cb2949ca4cd852",
      build_file_content = packer_build_file_content,
    )

