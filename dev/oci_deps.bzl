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
        digest = "sha256:ddac091c8c7aa8672c3ad097fc0df016f5cf86a2e5e5e915a5e6f0327b2cc48e",
        image = "index.docker.io/sourcegraph/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:d20378947f0b7df72ee6166a70ccf3839c7547dfe918a02f920ee5d4cb031f62",
        image = "index.docker.io/sourcegraph/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:d699c86891e976b68a7e4611951a90400805db552e222750e04cf3f846e09a2a",
        image = "index.docker.io/sourcegraph/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:c842da00bc3a7056c49902406f9996e4ef874066f80f06c502beb4dc716792cf",
        image = "index.docker.io/sourcegraph/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:a16a784a3147b768042b5496a08da429b01a5c81442d2f744216e6ee63c0a6e5",
        image = "index.docker.io/sourcegraph/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:461ef0c0233c20716398a0ad1a38dc2df955e26ec091400f64a18fa7de9d2621",
        image = "us.gcr.io/sourcegraph-dev/wolfi-grafana",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:109a393c6bc3e5db15413ddb913890fdd26f0531a225ff12aedf5021f57c5bd0",
        image = "index.docker.io/sourcegraph/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:0be08d4aa088c945f1cbdf6d69a618d3b803786e21d540d24d0ca6a2f7a9a044",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:7486c7418c6905f84cc95b2bb58ab3fa0779033d5da3045217ecf96c211bed7d",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:8ecedf267d9d017f7defe0ce9de7596b6e0f8d3dc434e4e54069d07a7f1d6f2f",
        image = "index.docker.io/sourcegraph/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:b60ee6c59823598a4bf6ff9e97a887947cb7b1a925d67f30e0448adcd540743b",
        image = "index.docker.io/sourcegraph/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:874544e9e2ac04ac3e7004437b55882b96e4f9b444f81d359afb89361c01c8a9",
        image = "index.docker.io/sourcegraph/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:26c41915424166a79b4e622d2b804ef28409df9d10123b05722467911e15550b",
        image = "index.docker.io/sourcegraph/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:69890df98e79031d3cb8bdb1b91d25eec585d987d9c3a77b7b2034f5f233107b",
        image = "index.docker.io/sourcegraph/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:b4a4b8a0afe40c514b62ab7e9f0e4b499f193e0ea7d22dc3b702b34ad2781799",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:82b3d39633915ddff0ba45a9c04eb627e10d0b2e698ede36386240faa91c5a29",
        image = "index.docker.io/sourcegraph/wolfi-executor-base",
    )

    # ???
    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:414df1099deb034bd46aa10432a2571fa342b129a3e59fde20c2bd63e84c4ca2",
        image = "index.docker.io/sourcegraph/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:ceb89c7b756b0134a8d512be72276e7d536921fb2619117059970c5d252389fd",
        image = "index.docker.io/sourcegraph/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:bc26a76ce432e2d2263024d7e285ef9a28801fba0e23f2c1f71d56d6a30bc2a6",
        image = "index.docker.io/sourcegraph/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:75ea7711988255d037e6d6239d915d1d2594452fc343b0b98bfec014c6b6fa9a",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_prometheus_gcp_base",
        digest = "sha256:be32b3432ab03a153461bee8c604b23413cf28c978556c30a483f8630f491b82",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-gcp-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:94cc57d7d62443784dc113ca6fdf072f4b3e58ec3636779b59dc9f96e238d56a",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:4915ea39e5ab816f935f85cb7e2ace2824d4aade17525f8d9d2863f27f8f27e7",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:a0615019fe6961519c994cfe0de2b595571f9b79a3e60a58b9d452279bcee79a",
        image = "index.docker.io/sourcegraph/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:d23cae6f2fe2db4c229c3db5dcbee25dac95de8b6d8c857d409fbc7c797f660d",
        image = "index.docker.io/sourcegraph/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:b4a4b8a0afe40c514b62ab7e9f0e4b499f193e0ea7d22dc3b702b34ad2781799",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:afec08789b60252c51163227e416a40deec5f34bce7b5db43b3f0647f7691ba5",
        image = "index.docker.io/sourcegraph/wolfi-blobstore-base",
    )

    oci_pull(
        name = "wolfi_qdrant_base",
        digest = "sha256:1d9cb0e1eef19467bc7b42ef36f40571eb37abb5f6eeb04742717cc837050ae0",
        image = "index.docker.io/sourcegraph/wolfi-qdrant-base",
    )
