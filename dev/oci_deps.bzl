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
        digest = "sha256:818eb895261404efd4952f11faa70e8530079f7b9ff9225aca74b343f49e3918",
        image = "index.docker.io/sourcegraph/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:930e29cf754a780e9948b9da793de99ec9ae94513ea80d917c6a67ac979fc855",
        image = "index.docker.io/sourcegraph/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:d82045bda06fc745f3a0bcc20507912f530f5d48fbb16f23d88cdfebcd7b324d",
        image = "index.docker.io/sourcegraph/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:269ddcfff8b268af2a348c25f4bf15d07951281fabff347a7c92bf3bb09cf65f",
        image = "index.docker.io/sourcegraph/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:b5dd7bfde3df1bc8b4f5a9134bf509f4facc172d721fd3d24ced22e28ada956d",
        image = "index.docker.io/sourcegraph/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:07a31e9c49300872deeb162910207611261475634bd7bc92e4dcf9a285c98f16",
        image = "us.gcr.io/sourcegraph-dev/wolfi-grafana",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:c5b1536f1dbb378b9f1da1f0b03ba3cff784965372c21943c902276f0b67dcb7",
        image = "index.docker.io/sourcegraph/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:c5546f349b20c02ebd7428a11e2a6e7a9ae6f89564327a64e1544a0dfbe8915c",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:d2e7110fb2deef97562bc472e6eab126e885f0229ad8162d7a96b2e05b74ca1d",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:c2ff8b3e0f21b5698c2a693741a5bdccaf0198908093eba348c9e21edbe1f630",
        image = "index.docker.io/sourcegraph/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:5abd3fd25651b5af3c30d25f33215a7e7d2c4108b6e28b0e6380879d831df68f",
        image = "index.docker.io/sourcegraph/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:f9cc9923f84989639b3d03ef3f8120deb9c2d3cc6d1e8bfc703fa9a5327b4612",
        image = "index.docker.io/sourcegraph/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:782e33b097b9f68ade8e613c21cc1531a8fc01208c021fe307b05cbdf6dca761",
        image = "index.docker.io/sourcegraph/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:5596447f4c22b5b305016f0fb095e3fc28700d364a660a635f03b14feeda5c03",
        image = "index.docker.io/sourcegraph/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:8ae66efec501aca7d6bc66a2b3a19775e2cd2b8db0e4e8c077bca9c764d55673",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:8f6728cadb10d37df4f8da1eecf59b93625977c974ff3d6e4747763eaa734cd5",
        image = "index.docker.io/sourcegraph/wolfi-executor-base",
    )

    # ???
    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:aa2816aedaf4964ed9263960bcbe62f5ce6c626e626d351d2073c281c61c4bad",
        image = "index.docker.io/sourcegraph/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:183442440539edc342eb7a136b4be35cb43ff819eb1b88e21e7db4ac5a6b373e",
        image = "index.docker.io/sourcegraph/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:a290fbe16018386fa96666f828cc983533cf6760030104e698a254fff5430519",
        image = "index.docker.io/sourcegraph/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:845a876fcbe7609cfa999517996bd7c1974c403f4cad81ff6e1e0c2a8cf2263c",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_prometheus_gcp_base",
        digest = "sha256:9e02d2c560539ae860bf08ec18748559cb78e11c71ba7342cf4c8816ce1604d8",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-gcp-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:24dc2c4153f12aea87d6abc085fd32c83c864ce04c95553b0cb402811e0ef517",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:81007215d3808d69658e4bf2305599c27fdffc5df9df1d0e3091f4ca062a2abb",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:15d2aa75d3c0d27de62109ea6245f1a7a0a9abf9bf0a97b19ed545b3ab911624",
        image = "index.docker.io/sourcegraph/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:fb77c37fe80a7e23796a5f0b54478dd075fb7ebf82a58a145d48aad0b03c2618",
        image = "index.docker.io/sourcegraph/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:8ae66efec501aca7d6bc66a2b3a19775e2cd2b8db0e4e8c077bca9c764d55673",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:41fe629da39e876f54902b072470ee379fcf5c3bf7340fc3ae6d294959f0b5bf",
        image = "index.docker.io/sourcegraph/wolfi-blobstore-base",
    )

    oci_pull(
        name = "wolfi_qdrant_base",
        digest = "sha256:c6d5509a1bfee07ae27d45347ff36a4cb0c14b23c7d597e161385c34b9742490",
        image = "index.docker.io/sourcegraph/wolfi-qdrant-base",
    )
