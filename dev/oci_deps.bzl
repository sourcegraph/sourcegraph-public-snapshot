load("@rules_oci//oci:pull.bzl", "oci_pull")

# Quick script to get the latest tags for each of the base images:
#
# grep 'image = ' ./dev/oci_deps.bzl | while read -r str ; do
#   str_no_spaces="${str#"${str%%[![:space:]]*}"}"  # remove leading spaces
#   url="${str_no_spaces#*\"}"  # remove prefix until first quote
#   url="${url%%\"*}"  # remove suffix from first quote
#
#   IMAGE_DETAILS=$(gcloud container images list-tags $url --limit=1 --sort-by=~timestamp --format=json)
#   TAG=$(echo $IMAGE_DETAILS | jq -r '.[0].tags[0]')
#   DIGEST=$(echo $IMAGE_DETAILS | jq -r '.[0].digest')
#
#   echo $url
#   echo $DIGEST
# done

def oci_deps():
    oci_pull(
        name = "wolfi_base",
        digest = "sha256:0ccfd730a1918144881d5cb312874ae20bac84bdf4a90613e7b433423a40161c",
        image = "us.gcr.io/sourcegraph-dev/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:afb1c1179db7f89114dfc402495c34b3cf79a53e914a6f85fdd383d1808a570e",
        image = "us.gcr.io/sourcegraph-dev/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:2472c260bd024dd6d92bd1f3dddd0783f2e9e26e2cd325f74cb8c97f279b1ce0",
        image = "us.gcr.io/sourcegraph-dev/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:0cded0aabc7509d0ff02cc16f76374730e834d8aadd875f47bb4dc381d4105a4",
        image = "us.gcr.io/sourcegraph-dev/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:ac49e90f580bd7f729c78116e384feb25e8c2e1dc7f1ed69395901e68cfd82d1",
        image = "us.gcr.io/sourcegraph-dev/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:ec1049f35ff7e4ab6ff7b4cc6790996ad74d196b8dcee8ea5283fca759156637",
        image = "us.gcr.io/sourcegraph-dev/wolfi-grafana",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:f71d13c7b1a687a61a3954c11005b4d65773d0d857e8622d846ab83a6b29977f",
        image = "us.gcr.io/sourcegraph-dev/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:f8e416626dcdc7d14894876e0751f2bfa0653169b14c4a929b3834d30cea087d",
        image = "us.gcr.io/sourcegraph-dev/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:77da62533456112d87f61b24d6694b2bb7276446e6f94642580cf9649641c4ed",
        image = "us.gcr.io/sourcegraph-dev/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:08e80c858fe3ef9b5ffd1c4194a771b6fd45f9831ad40dad3b5f5b53af880582",
        image = "us.gcr.io/sourcegraph-dev/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:003f8d2411cf198de0fe3751cef7b08ed85f2ff05746097bee9cbcae304eca31",
        image = "us.gcr.io/sourcegraph-dev/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:9c4531f4f0263ad49678ff81f1094d2eb5f9b812bd93e21da19780b480fe7c52",
        image = "us.gcr.io/sourcegraph-dev/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:c30c2bc85aa38c9c3796038b290f3eb8fbea6b3e744f91788860de4e58bca822",
        image = "us.gcr.io/sourcegraph-dev/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:19445a121968c19bcd3bf5ae05cc97802853c039d00efc83f317655def510170",
        image = "us.gcr.io/sourcegraph-dev/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:2d90d34644c473bcf5f998c4ce881354992bc28d0644d47e182c2475f0bb616a",
        image = "us.gcr.io/sourcegraph-dev/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:e73c7c629307a2668b5a199bba79d315e7bd8df414e27399e723a8630c06c08c",
        image = "us.gcr.io/sourcegraph-dev/wolfi-executor-base",
    )

    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:a1229fd8a3511c6931293f3a7b22974741f8def858b54836590a488064cf8240",
        image = "us.gcr.io/sourcegraph-dev/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:8e55b6529c84bb0ff24f2d8dc889b74263bcb2584312028ba70d4ce9147d10c1",
        image = "us.gcr.io/sourcegraph-dev/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:68dfe2e32c698f457d4f096baf13d7c052b65f80ed3163a15ab30dd2836daa88",
        image = "us.gcr.io/sourcegraph-dev/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:fca22248cf10c90af2445cc6627d0300dd46a23be89afd0899618f896909feaa",
        image = "us.gcr.io/sourcegraph-dev/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:cf8a07db3ad8c439e85d4142b3d3f3ef394551a9077b41aea0788f4979c0a9d3",
        image = "us.gcr.io/sourcegraph-dev/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:6229efd204ae3869fc5f0441da54dd0d4864a972f0ebb0f1e514a18d5fdee0e8",
        image = "us.gcr.io/sourcegraph-dev/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:4bdab308f1c538b3df4d0c2b52a1b47e56ba4fde2c5d8083e847ee86ed8f7320",
        image = "us.gcr.io/sourcegraph-dev/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:e43edee16894ed6c94fc9505e7387958bad929f33842a1e7c7dba2a0fcf50aa9",
        image = "us.gcr.io/sourcegraph-dev/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:2d90d34644c473bcf5f998c4ce881354992bc28d0644d47e182c2475f0bb616a",
        image = "us.gcr.io/sourcegraph-dev/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:75f055320da219bfd83695bcaa011a93c9e101f00b6100b70b04a0e03bd661a3",
        image = "us.gcr.io/sourcegraph-dev/wolfi-blobstore-base",
    )
