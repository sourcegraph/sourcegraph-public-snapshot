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
        digest = "sha256:a236182c8e16e23aafc2d96a0baea17414626064373e38b9abbd5056c3c4d990",
        image = "us.gcr.io/sourcegraph-dev/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:11ab7cbe533a01b31d8b8803ec8e6e52074f6d5ba20a96166f7f765db4e8819b",
        image = "us.gcr.io/sourcegraph-dev/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:c1b84bd2c2840bbed25a65fa6ade38755be180f4ca0b60e45b461339073a5dc6",
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
        digest = "sha256:544d4f8a44cd03c7110855654d17490d3d485d69198a8789a2bfa25029a66e09",
        image = "us.gcr.io/sourcegraph-dev/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:f8e416626dcdc7d14894876e0751f2bfa0653169b14c4a929b3834d30cea087d",
        image = "us.gcr.io/sourcegraph-dev/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:aa6ee947778196115d3027eab91bf0d0e0cc91a03d01f5c2854cd6ecf97b089f",
        image = "us.gcr.io/sourcegraph-dev/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:4945cd8307f1d835b9b9607a1168ecfb84cdc5a5c14eb7c4ba84c08c50741b7b",
        image = "us.gcr.io/sourcegraph-dev/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:d1a3d302a4be9447f2b0863587cb042e69c5cceb4eaac5294c9632b58f285a64",
        image = "us.gcr.io/sourcegraph-dev/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:0c777bb76c4e586702f5367f54e62881f2f0fa5a96a1bd519ebaff1e982d1ef1",
        image = "us.gcr.io/sourcegraph-dev/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:991f696b62c4afa2ced41b1071b15e44094a2c541d973a831b34d4a4db4c2131",
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
        name = "wolfi_bundled_executor_base",
        digest = "sha256:8941bfcf8db44c462c4fee057b4d2a128cd268d6fb385989086af4381a05137e",
        image = "us.gcr.io/sourcegraph-dev/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:c2053b17cb8904a09773552049929b73215af24520ca22613cb5b16f96d8bcfa",
        image = "us.gcr.io/sourcegraph-dev/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:2abe940a2d9e13a998d07e4faf072c7ba6e17243a0b9c56a3adf9878d9332f6a",
        image = "us.gcr.io/sourcegraph-dev/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:11a84d7ae6a7f3a8954306f224391a2301e5518fbe2d6e8dfee4abe12ca91180",
        image = "us.gcr.io/sourcegraph-dev/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:7c8bfb96038fb6b3980fb6b12f692a54ff8f1d1cebbd72d9706578be8d278cae",
        image = "us.gcr.io/sourcegraph-dev/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:4d85ed245a6d3f22a42cf2fdf840a8e14590fb5216f597a19697e7f3025a0a26",
        image = "us.gcr.io/sourcegraph-dev/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:0a42810eafc6c81f95cb3295003de966d3a857d37dd79ef45a3be54c3b4c8e7c",
        image = "us.gcr.io/sourcegraph-dev/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:d5d55fb77056422eea328b709795a38cf599e42f2a90787c9dd32c2f1a3654f3",
        image = "us.gcr.io/sourcegraph-dev/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:7ae7d14bc055f5dbcdc727261025e9527558513a61093e956cb39dae7dbc0dcf",
        image = "us.gcr.io/sourcegraph-dev/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:4076564aaa3bfc17a8e50822e9937f240f155a69f9cffcb039a49f892446d28c",
        image = "us.gcr.io/sourcegraph-dev/wolfi-blobstore-base",
    )
