"""
Load external dependencies for base images
"""

load("@rules_oci//oci:pull.bzl", "oci_pull")

# Quick script to get the latest tags for each of the base images from GCR:
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
#
#
# Quick script to get the latest tags for each of the base images from Dockerhub:
# grep 'image = ' ./dev/oci_deps.bzl | while read -r str ; do
#   str_no_spaces="${str#"${str%%[![:space:]]*}"}"  # remove leading spaces
#   url="${str_no_spaces#*\"}"  # remove prefix until first quote
#   url="${url%%\"*}"  # remove suffix from first quote

#     TOKEN=$(curl -s "https://auth.docker.io/token?service=registry.docker.io&scope=repository:${url}:pull" | jq -r .token)

#   DIGEST=$(curl -I -s -H "Authorization: Bearer $TOKEN" -H "Accept: application/vnd.docker.distribution.manifest.v2+json" \
#     https://registry-1.docker.io/v2/${url}/manifests/latest \
#     | grep -i Docker-Content-Digest | awk '{print $2}')

#   echo -e "$url\n$DIGEST\n\n"
# done

def oci_deps():
    """
    The image definitions and their digests
    """
    oci_pull(
        name = "wolfi_base",
        digest = "sha256:7765ccf698d40dd46b1b69599728680313ac993e3b8ee2292faf99f9acfadfae",
        image = "index.docker.io/sourcegraph/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:17ff4201bf1904e679017d08c0d0216a8b10f85e096c64f26efaf73006849146",
        image = "index.docker.io/sourcegraph/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:2972e10f5dca979930a5b1bd2a9cd63eb277e9b95a77a00a8c2e63d52022c1d8",
        image = "index.docker.io/sourcegraph/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:259ddda8d0d4297fe4231c017b39b90dfa0fb8fcf28aca61f69fddba16fd8474",
        image = "index.docker.io/sourcegraph/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:5fa1bf66f3f90791ecbd28ca51dac748d72c37abdc49ba63b728258414128c61",
        image = "index.docker.io/sourcegraph/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:4644fa597db008b020f8cbc620ed365854dcfd464832b24bc033b2ce067545ec",
        image = "us.gcr.io/sourcegraph-dev/wolfi-grafana",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:c22f0a99899c29a6bdd3d59f7c1c83d5aff180a0baba6bca4c2f28950da952a9",
        image = "index.docker.io/sourcegraph/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:080ede778a1a7e009a45a17ce60ace5431df8e589ce611d8f52033024cc64410",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:4e1ffceab92819d6eafa785c4e2b00da1930c89361888623a735a1d317a0c74c",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:143082fa39dc9cd510d5c9011fafceaefe05ea194098d71ec3e61189660b39fe",
        image = "index.docker.io/sourcegraph/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:1caa8489778b89a8bf326f4092f37886b6e454243a2771b792d0ab086be2deab",
        image = "index.docker.io/sourcegraph/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:ed78d0d25c93c8c67742214a7a07a9e2e33d27727af6f6c0a234159bccc919f7",
        image = "index.docker.io/sourcegraph/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:337282a47ae438a2a4124c23e65e7d5953354dcb31a6c5d53fde6001df237e2b",
        image = "index.docker.io/sourcegraph/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:9589a1c4ecc573b0f126beb46a8418fd14c7add9e11db839d7f7bfc24b36ec3f",
        image = "index.docker.io/sourcegraph/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:d05443bbeba35b46b93e7d69377eb1789e22df5ec0e019497b8f8441bae1c186",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:2826c9fd3a0e33cfbc8df2dca3ad53c31bd6a5ed921144e2f93f41ac8663ece1",
        image = "index.docker.io/sourcegraph/wolfi-executor-base",
    )

    # ???
    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:b22f6175da2d6124183bc39e41adac8cc106ab2d0d71b141ef0bcd84185b1f52",
        image = "index.docker.io/sourcegraph/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:a5d60c3a889a0d99b41c241d56f309db00db3cda6cfd5bf145706c18b661b29f",
        image = "index.docker.io/sourcegraph/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:184eba7fd8e298a81f6e44895a4e64fe14b5bf0240018ccc2997c410abd9f63f",
        image = "index.docker.io/sourcegraph/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:8c497e05ed35fa07de08b558f2641389dd82168845dbbda8ffdfaa843b1ac6ab",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_prometheus_gcp_base",
        digest = "sha256:c9b1c8f4b62588b924ae791a4948186c0fb8029c7bfa4c925df69cb0561bb4fe",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-gcp-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:ba71a8c994c7b6e677a0cc4d897309ed93f38bc26e829fac3ee1f28eb3d8b898",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:779ec1695a677d633c1adf43a8e137ba475754809ef07835cf3ecfad23a3c868",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:4563d813a54a5da0dd2c4ce12e62e24a1c5372d557a6684e6b2333ecb4f6ee17",
        image = "index.docker.io/sourcegraph/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:438052c65d55f04138a45bfbadd399d0412c914ac7059cf068e84a6edf4fbe42",
        image = "index.docker.io/sourcegraph/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:d05443bbeba35b46b93e7d69377eb1789e22df5ec0e019497b8f8441bae1c186",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:cf83ed9f5c274d756d30149c922836b86902d8b0cb7a50f6e2f57a1cccc80d56",
        image = "index.docker.io/sourcegraph/wolfi-blobstore-base",
    )

    oci_pull(
        name = "wolfi_qdrant_base",
        digest = "sha256:959adb714604ee9908724e6da9da620c0ddbabba772cd6448d2d35ffa7474adb",
        image = "index.docker.io/sourcegraph/wolfi-qdrant-base",
    )
