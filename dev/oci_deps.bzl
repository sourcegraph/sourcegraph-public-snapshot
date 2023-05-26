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
        digest = "sha256:3d0d41407333ab690c4a6d1905b7917c5c867ed693a2890e24b85ebafe3f9d8a",
        image = "us.gcr.io/sourcegraph-dev/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:8de4ec6b220a32e6e4348980c9990c46ce0d5d6294327421768dfc07311a008d",
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
        digest = "sha256:2473c26cc21777f8051ac86d10ac680b0d06071f2949cc2e300a6904d5235d5b",
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
        name = "wolfi_batcheshelper_base",
        digest = "sha256:2abe940a2d9e13a998d07e4faf072c7ba6e17243a0b9c56a3adf9878d9332f6a",
        image = "us.gcr.io/sourcegraph-dev/wolfi-batcheshelper-base",
    )
