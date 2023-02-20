load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

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
      sha256 = "633f242246ee0a866c5f5df25cbf61b6af0d5e143555aca32950059cf13d91e0",
      executable = True,
    )
    http_file(
      name = "shfmt_darwin_arm64",
      urls = ["https://github.com/mvdan/sh/releases/download/v3.6.0/shfmt_v3.6.0_darwin_arm64"],
      sha256 = "633f242246ee0a866c5f5df25cbf61b6af0d5e143555aca32950059cf13d91e0",
      executable = True,
    )

