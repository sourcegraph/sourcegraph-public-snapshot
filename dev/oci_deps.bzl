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
        digest = "sha256:ae7394b38d185569fa082bec69b4393d23eff70a8fb93590a9d07e4f17bd4106",
        image = "us.gcr.io/sourcegraph-dev/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_base",
        digest = "sha256:6e2200f85c7a7cf6a831f2f51082756b6ba9107c261ad00a9b7d2ed65e4868c3",
        image = "us.gcr.io/sourcegraph-dev/wolfi-jaeger-base",
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
        name = "wolfi_prometheus_base",
        digest = "sha256:11a84d7ae6a7f3a8954306f224391a2301e5518fbe2d6e8dfee4abe12ca91180",
        image = "us.gcr.io/sourcegraph-dev/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:b1e08b459ca4cf7529ced3a611361834859014ab508ba0b53914f14c9fc269dc",
        image = "us.gcr.io/sourcegraph-dev/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:c9342cdec5a5889343e816f9ffd0707a50cbfba084339d0649bbfaefc1ab546f",
        image = "us.gcr.io/sourcegraph-dev/wolfi-postgresql-12-codeinsights-base",
    )
