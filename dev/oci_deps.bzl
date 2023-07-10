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
        digest = "sha256:787447d9103e52faf924e7fcd71bb052f5cecb99b75ec6cd2f37453405781bd3",
        image = "index.docker.io/sourcegraph/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:42e9626596c7e3873f9ea66b87732273bc93f20d05499d410160468f40d585f5",
        image = "index.docker.io/sourcegraph/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:4bc66d4d457f0d32ff359389bf4da1d3b6864043a7ee2e679f80a8d0d5daae32",
        image = "index.docker.io/sourcegraph/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:8c43c861026ff3706849bf56d22ef2a1714f76e739fb2cd9f89b78abeec39bd2",
        image = "index.docker.io/sourcegraph/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:37100ed0a90e13f0a84fb7e05b49e75700ebb0b036612f36f5afca69fbcbf86f",
        image = "index.docker.io/sourcegraph/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:1989633b60c45edf330f1161f7be37bc75b905b0947607e175bb82db224ce50a",
        image = "us.gcr.io/sourcegraph-dev/wolfi-grafana",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:df37c2ab0fac3e0215b19f51cf993971a71db9fb4db1230070e937d7842f154b",
        image = "index.docker.io/sourcegraph/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:f5df06fba21785e56a345bcc8ff4d197c9d5bd95392119447dcf6633e35746a2",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:8a26bb610eb28d55119156e594db18e72c80c893971caee3f973c9ed6626df3b",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:12deb901ebeb581e0fc1b175f63235e1915795e1821d0f06dcde0c2030018d17",
        image = "index.docker.io/sourcegraph/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:540db7c4d1a033d7cfc242e6c7a87ae4682459e3c82d99fe6f8d04a004e95367",
        image = "index.docker.io/sourcegraph/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:a637dc7a1045e6324048c35b8d2dd427d421304eaba64b486581f0a1d1f3e0c6",
        image = "index.docker.io/sourcegraph/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:bdd50db38bf7562c142f8d88a618c9a9b4bcb060e1c374a7fcd8bef33c06a831",
        image = "index.docker.io/sourcegraph/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:2a891f3d48272b4e4adc7a7aa7f9597c18c2685ce5ba3a42f8d6745e9e3eeac5",
        image = "index.docker.io/sourcegraph/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:e7a2efcd084da4057fed5ba342fab6a58f2e20cfb839d383729786ab1a00942a",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:7ff7c9030d39e7345d6c8479f5c7a27818f28ddd76a195f30af50bfa5362388b",
        image = "index.docker.io/sourcegraph/wolfi-executor-base",
    )

    # ???
    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:7418c2c8a4da7a044f2e16e6e36f730fe3d3ad41e9af32f1bb471c2026b613c3",
        image = "index.docker.io/sourcegraph/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:4afb06ca2d4c4585fecb0443546b7f61f4b88bf8b4f74cb005e712e3e76dda21",
        image = "index.docker.io/sourcegraph/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:2572c8ed4a5286fab501c209e88261fe21347fca6b749f90eea3d90a768f894b",
        image = "index.docker.io/sourcegraph/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:cc529187858e9e019a6ff2fdb5da4f5cff2d8222987bdc59b386e53eb1fa1a63",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_prometheus_gcp_base",
        digest = "sha256:9981265b4e0132176bb3344913a1d5eda3f4fb77560028fe2e076a9d505a0a1f",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-gcp-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:e3272de2b6958aa64d7429aa2251470c4adab777496bb7290fa1a51d53193e7d",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:a242c9a88552a7af2d70f4be835267cf89e4639f8607dea7a9f98c2909136734",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:85b5b0f1cf4df378ac007b846a84fde3a5b00696394bfa2dbc6ca0118ba31999",
        image = "index.docker.io/sourcegraph/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:e6ebf3853f5ac70f7d63d35197a1e2ffb802c0229034e68b46c608d74e27dd44",
        image = "index.docker.io/sourcegraph/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:e7a2efcd084da4057fed5ba342fab6a58f2e20cfb839d383729786ab1a00942a",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:7e9e3e471fe8a3483e86a3c246b023684b20ebe249dc6ea0c5a9fef1bced5395",
        image = "index.docker.io/sourcegraph/wolfi-blobstore-base",
    )
