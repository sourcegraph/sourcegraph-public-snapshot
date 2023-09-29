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
    oci_pull(
        name = "wolfi_base",
        digest = "sha256:f7d19d251a65471e5c96114823b480fa6ee2649e4d30041b8a75bf84fbdc9293",
        image = "index.docker.io/sourcegraph/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:c98bb71adef2e6412ba6b15a8d75e9cd2aa761238458afbc08011a66fd73a536",
        image = "index.docker.io/sourcegraph/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:e1707ffae8627bfabb972a31bc2d08973c8b05f344812d6095f670957c27a503",
        image = "index.docker.io/sourcegraph/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:79fa14df95a902dd5f9d3c5391cc46664277e3ce8e3c7d1f2f3059066af55f6f",
        image = "index.docker.io/sourcegraph/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:1b63c61ff9d704d1f271e9b480ee63a40da3933b06852870f251a16141c39f7e",
        image = "index.docker.io/sourcegraph/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:6f90153789f26885fd708ce695a8f83959216496f55ed9bf9dfe26288a7f1f41",
        image = "us.gcr.io/sourcegraph-dev/wolfi-grafana",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:d4dc495c3724b7035a42563c6af1287f3bd50c7c38157555ae82c4990fc17433",
        image = "index.docker.io/sourcegraph/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:bc045404452e77f29be50c127ed6747e6e036d3dd2c0fe4def0e9dd3b033f887",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:f3d3860d57d4371c578ff8f4a828b4834f65b2149dafff76821416051842671e",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:5521667b1f3b196851eb802923db563ebdd1834d2ba079f5b80225369c849e8e",
        image = "index.docker.io/sourcegraph/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:3c56394e8e307d56937601b3fc6835aa70e9c135ae2602794f1df460dca07879",
        image = "index.docker.io/sourcegraph/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:8281dda4aeee958da686f6f0b12ae7fdeec964a2cc5c87f2c26c63fcc323201c",
        image = "index.docker.io/sourcegraph/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:6a559d14325ebb6a322fc7153e4e2320a4959dcbba65e3a203a25363bec162bc",
        image = "index.docker.io/sourcegraph/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:e66998c123cd46de5121b29c0badc24c8e51283b553b23465a709d3d3e467691",
        image = "index.docker.io/sourcegraph/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:e7ea5dfdcd4d9d272727a7a984f020f5ee1be30eed98d498bf30e5a89f9d47de",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:ecd37a5f55bca3bbf090bc1402f49a8ffc1935007b76a9d896b1a728bb665301",
        image = "index.docker.io/sourcegraph/wolfi-executor-base",
    )

    # ???
    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:6e8f04b1fe43e89f8978c7b24e1c9b45879e69005dac59a881fba5d04d780860",
        image = "index.docker.io/sourcegraph/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:406b30df34df4de6276d28d46d8026dc5c253cabbca5354032fd5f8134136de0",
        image = "index.docker.io/sourcegraph/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:7e73e28af8537559a10ddd7a6257f39c62ccb94c596c3c2969e5184bca8e4c7a",
        image = "index.docker.io/sourcegraph/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:8caa94d4c935c9a62b03f6ae5f731badc4a4eab9f4e5a0bf53ac5605a8c40439",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_prometheus_gcp_base",
        digest = "sha256:40b3ea43e94266f61d086c05cde5ef46c4a7a6317fc5862455133cbf4b7afb65",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-gcp-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:60b6197f9ecb38de089b64e6c44966809b4b5e659be527369dec86cc2070b6c4",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:6c77b2d4c08850814969aa7232bb3b3c65c7f560abeee27c3558fdda3b210bae",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:6ccdbd6e573b17a701eaf458a1e5608406a0d227c2e8d8e13426393991c05eda",
        image = "index.docker.io/sourcegraph/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:eccb80a850a322b064cbdfcb34a36b8e67e8ead7e5fdff9a209449e029937985",
        image = "index.docker.io/sourcegraph/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:e7ea5dfdcd4d9d272727a7a984f020f5ee1be30eed98d498bf30e5a89f9d47de",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:db975d47101afc417ab434d7869599b08103438df9744967d156361b43543e46",
        image = "index.docker.io/sourcegraph/wolfi-blobstore-base",
    )

    oci_pull(
        name = "wolfi_qdrant_base",
        digest = "sha256:3ff70f00bb5ffcb0521e89b92c9b746b0d25ce778c4a0f00a3a24407e7c466e1",
        image = "index.docker.io/sourcegraph/wolfi-qdrant-base",
    )
