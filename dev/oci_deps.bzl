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
        digest = "sha256:6ed590d3adebe3bb5a655f672fa8a599ee7b432e18f0831227eb46fa3cd29dae",
        image = "index.docker.io/sourcegraph/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:442a0eb9bc61c809b7e26f21d214bf0c5923c89803485d4ab123db28ba77b537",
        image = "index.docker.io/sourcegraph/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:6c5ff11cbe30fa87220573f95d13cf12f4ea432c33a89a8a31f4487fd0006a22",
        image = "index.docker.io/sourcegraph/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:a1147fd551868ec84dafce9ef1d2f9c6a64a685bf55e0f4034d4a6f3ed6ac388",
        image = "index.docker.io/sourcegraph/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:938ef779926a88c15a9dde2fdf77a18085af8ea3fbabb12746798f10bd0627e3",
        image = "index.docker.io/sourcegraph/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:963ec6435d899ec8312fd3555f96f117e5cd018218117bb6f89c9313e53ae6a2",
        image = "us.gcr.io/sourcegraph-dev/wolfi-grafana",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:2f04939a502f98bdc239d914f98d8e308cbdf0ebde05daa5b3c78791f8e92d28",
        image = "index.docker.io/sourcegraph/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:cad0d67abccce789e8bb53daec5e5b85b392a7ef952454250694dddf005fa1e8",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:b2992e8a7fdb783da50db91249b1a1115a6f4e667eb8bffd9a71a1f805bef346",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:4d6eea2f193d72f8d4ea4f21b5dd7f80aa57204fbadffeffe40e76af36637ac5",
        image = "index.docker.io/sourcegraph/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:479528edc55bab31cd6f98b0f90562fd95600bac227e911abff1cb27bd5ea0d8",
        image = "index.docker.io/sourcegraph/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:d100bb61f21d9217fe2bceca1b60735d2d1859a7306734dc6b628b993a86d2cb",
        image = "index.docker.io/sourcegraph/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:c31bf4cdf7a7b52e9697d941f1a83cbc4d82b792dcaa991896ddd6e842ac106e",
        image = "index.docker.io/sourcegraph/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:f3d2dfa1c32d5ccecca29f9595eb062f0ddd815418b6904a35a78b5d34f4d54d",
        image = "index.docker.io/sourcegraph/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:88cf6039d37ebfa79f3dee6f2862aa8441d3e65a14cdf71214d32d04fc8c5c85",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:c030c215650b39dabb9fc5cb7d438431a936f59dcae559d1751aa1e9f603ce52",
        image = "index.docker.io/sourcegraph/wolfi-executor-base",
    )

    # ???
    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:aa931c25e13971de892b60e0ad3d892cd6eeabdaf2dac6d6473b238eddb85513",
        image = "index.docker.io/sourcegraph/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:89e2f5fcf1c69b02254a2ecc4857beda7288351bd52b8fb1e88c1c8f0182995e",
        image = "index.docker.io/sourcegraph/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:72a97d66c9f701aae0dbf3a5ea6330f209fe2240d5a0018c76b8e9b9d2d45dad",
        image = "index.docker.io/sourcegraph/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:c021e04d20af9a79b4ba2eedb7c83bf78e259a99e43d34754c3d5bda4add445c",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_prometheus_gcp_base",
        digest = "sha256:9835899e66e57c24eeb0c6e92b452d46c5fa8287fe85d1718b725c647604fb28",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-gcp-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:818933fff05a6f9b6f8ef738fabad7fcc70e57d873ddbc8187e3445a1e815aa5",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:d55b9886a1c2321a9f6e215ee3ce8468a28ebc638d9c224c5fafa4b7ade62912",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:72ce533b14be7f2ec75e32fe2063e6d65ef41d1202ac3e206b4411f3000b268a",
        image = "index.docker.io/sourcegraph/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:41a2fcd6cb01c4d0f7bb33f05fe8a1e3991e5eae38b145af0975c50df3e77677",
        image = "index.docker.io/sourcegraph/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:88cf6039d37ebfa79f3dee6f2862aa8441d3e65a14cdf71214d32d04fc8c5c85",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:27d983a210084f2ee2dc73686ad874958c33fd78a63c0c9fd12582842016ccdb",
        image = "index.docker.io/sourcegraph/wolfi-blobstore-base",
    )

    oci_pull(
        name = "wolfi_qdrant_base",
        digest = "sha256:046e2ac46f26cffa8f0ff71b293181bc2e3ba53afcd0e4a7cf8ffbd8dfa11eee",
        image = "index.docker.io/sourcegraph/wolfi-qdrant-base",
    )
