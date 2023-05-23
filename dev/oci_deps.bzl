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
