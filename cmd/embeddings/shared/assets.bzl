load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def embbedings_assets_deps():
    http_file(
        name = "github_com_sourcegraph_sourcegraph_embeddingindex",
        url = "https://storage.googleapis.com/buildkite_public_assets/github_com_sourcegraph_sourcegraph_cf360e12ff91b2fc199e75aef4ff6744.embeddingindex",
        sha256 = "830a3b4b05f889e442d1da0e97950136db907ffd395f5fef404d9b4a9aac84a7",
    )

    http_file(
        name = "query_embeddings_gob",
        url = "https://storage.googleapis.com/buildkite_public_assets/query_embeddings.gob",
        sha256 = "48e816d9ad033d2661a5c2999232ceccb430427ed4ce6330824f7c123db05a03",
    )

