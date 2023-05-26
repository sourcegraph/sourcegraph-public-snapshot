load("@rules_oci//oci:pull.bzl", "oci_pull")

def oci_deps():
    oci_pull(
        name = "wolfi_base",
        digest = "sha256:79d9c1e76beff31da0f182f30a2664dace9d9153cad8cbde7dba5edcef9e564d",
        image = "us.gcr.io/sourcegraph-dev/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:f9dfe8e6d8dede8cebc3fb4f4cefd6a78b93ad57b56e9da868ca59e8daadf7b1",
        image = "us.gcr.io/sourcegraph-dev/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:67731f797ebf8f6f1dcd08a5f4804adeee9ba2b7600d5ba8ed2329b64becd59a",
        image = "us.gcr.io/sourcegraph-dev/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:44f26735900d5319b23a139d48245ea009b4849d257ca53914b077d9430c1633",
        image = "us.gcr.io/sourcegraph-dev/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:045287c24f25e143d60997a502cf9c39addb8b815eb296fe48a02ca4f0ad9a18",
        image = "us.gcr.io/sourcegraph-dev/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:ec1049f35ff7e4ab6ff7b4cc6790996ad74d196b8dcee8ea5283fca759156637",
        image = "us.gcr.io/sourcegraph-dev/wolfi-grafana",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:d6941f29f96f3d94b10d7aa00e6dd738a96a186c62ef5981893680dc436c7fcf",
        image = "us.gcr.io/sourcegraph-dev/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:41089f3e2b3779681b585080e4748a3c7d5a5e086e19317ce4b943dacc2d9502",
        image = "us.gcr.io/sourcegraph-dev/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:e1d1cc439c12be48549a0ce4ea7a4fa3abb08c18ee94f73f60814d29c36e8e21",
        image = "us.gcr.io/sourcegraph-dev/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:935b6f7af308e235b48255f5ffa495daac3300e7366137294c0c8c90896a270d",
        image = "us.gcr.io/sourcegraph-dev/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:4c0072ee682f54ba793cd182fbe186a6a3c9672c245e04fc4f04cda192ca6dac",
        image = "us.gcr.io/sourcegraph-dev/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:7e5aef3d0b3e043761f83def2c1f761143659eca5738b4ba0ef1313e4a9b5bbd",
        image = "us.gcr.io/sourcegraph-dev/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:d3af56ce7dcaff85c2c48d3fcd2f2059365fb50beb8e545e08b8519ccc0e67e3",
        image = "us.gcr.io/sourcegraph-dev/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:74e478195b750c5547d6d240bc5d9e94b10d470d6bf2ef855dcd912f83550cdf",
        image = "us.gcr.io/sourcegraph-dev/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:fba8f4cce1306463b03c0388eb830bac63f02f4b8941c4df8f9fde99390da32e",
        image = "us.gcr.io/sourcegraph-dev/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:0ab096b0ffae9054fa18fa8121b105acfda767c5f74fd8530f72f8fe87ef20c2",
        image = "us.gcr.io/sourcegraph-dev/wolfi-executor-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:2abe940a2d9e13a998d07e4faf072c7ba6e17243a0b9c56a3adf9878d9332f6a",
        image = "us.gcr.io/sourcegraph-dev/wolfi-batcheshelper-base",
    )
