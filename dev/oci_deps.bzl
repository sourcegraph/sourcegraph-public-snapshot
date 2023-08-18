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
        digest = "sha256:0c318f31e2283606a8638d8f12887c61e1a638c169f3f893138f31e8994138dc",
        image = "index.docker.io/sourcegraph/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:461b071e88388d098ecf9229215f93300c65961c1998b15a5ea2e8355e50032b",
        image = "index.docker.io/sourcegraph/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:5ee69c6ef578acd7ebe3b5657fd0f7ab1bb7e67ddc3149c487a0304c9ff90a9a",
        image = "index.docker.io/sourcegraph/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:f169aa4f1dc9a6901a64ae78f195e8040b2b4207a3dfaf234862a963db9dda54",
        image = "index.docker.io/sourcegraph/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:811f6b8236ae26fc300c06700b5a49ab054f7a2637c9fabf3d71549d87a39fbb",
        image = "index.docker.io/sourcegraph/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:6f90153789f26885fd708ce695a8f83959216496f55ed9bf9dfe26288a7f1f41",
        image = "us.gcr.io/sourcegraph-dev/wolfi-grafana",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:85e95314e514bb1bdaec616d0fc073ccd8d1c586eb05af205e23debadf6e39fd",
        image = "index.docker.io/sourcegraph/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:b073b40468b8154c2caab57bea2a011b60d2a44410e6901c3cec924052337312",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:9cc8b58db182c460fc7a0cf6bb20fd89051e08c9d202181cd30b2b38c3f9d5b1",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:dec7c57abbd0061d23bb1013c2edbe7f2da5e57a4739b7418916669b4d644384",
        image = "index.docker.io/sourcegraph/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:3f314fbfec7b6604991eef42138ce4ee75868e4ee7f515f5b72b44a9592c4a10",
        image = "index.docker.io/sourcegraph/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:878ca09163885c6673f5acd89a1870748faa97e26ef5696297185d4bb2ba2d6c",
        image = "index.docker.io/sourcegraph/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:c3ef437358e36fd1d889997333817777cd67457d1c316ae23e8de5aaebc6afee",
        image = "index.docker.io/sourcegraph/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:ac23e9e74f5ad4f769ea386131832a00baf2d873a4c809d278373b2b9a1f8200",
        image = "index.docker.io/sourcegraph/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:184101e1c3aa4312bf359851c15ae04d8e5cc3441008499e92f784fe1c59a96d",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:7ccf7dbcf9c6b5d818504857d680bef1c0ac12674d2d61f80d4e45c7c07272c2",
        image = "index.docker.io/sourcegraph/wolfi-executor-base",
    )

    # ???
    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:7f01d45f05f39c8c5c1f9a9010ccd5e1f8ca7f70298ede88e37aa85052715ac2",
        image = "index.docker.io/sourcegraph/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:a11c399dd9880eeba82b22a03185f4896dcfb79963290d7105268f3e5485e7ae",
        image = "index.docker.io/sourcegraph/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:5d453c49f9a49daeabff155251f32c8f6e8157c8263dbe2a2839d45d64549800",
        image = "index.docker.io/sourcegraph/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:d336069d5846f068f00e9404fa64dd2cd36c49869ebc923cdac3d08c1913de3c",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_prometheus_gcp_base",
        digest = "sha256:14519040dfd547b933d6c345181f9a3a0eaeeb232c2922c7fbee52b05fbbbcdd",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-gcp-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:1e6c90c133107ef6655f14bdfe5331702465d9734c810c5c58c7ceef26565599",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:82502505d946eda47024f957ea308e7473e7f4ab5d19ca87826c352a683423d3",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:7a9367238123005c0f794362a012c8f50a84bb614a3a60f761a8c815262f2141",
        image = "index.docker.io/sourcegraph/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:52e6c51d0960614132da937907a74e844183bbc9f8c08b427135938469c6cf28",
        image = "index.docker.io/sourcegraph/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:184101e1c3aa4312bf359851c15ae04d8e5cc3441008499e92f784fe1c59a96d",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:42ff3ba38633022bfb2d7982e2ad671670b160fdaa04b453368f2098aaf9bf86",
        image = "index.docker.io/sourcegraph/wolfi-blobstore-base",
    )

    oci_pull(
        name = "wolfi_qdrant_base",
        digest = "sha256:7f81b52aa6accda3ccfd8d8ec79892de03b8935c7fb0e786a445de0bf1824a93",
        image = "index.docker.io/sourcegraph/wolfi-qdrant-base",
    )
