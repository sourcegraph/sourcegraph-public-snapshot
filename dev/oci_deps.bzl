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
        digest = "sha256:ec9bae746cb9a1dfd2c60ab208add8cfaca0e126a1d5a357377ce1440134b2dd",
        image = "index.docker.io/sourcegraph/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:083555935177510fe2335b23e08d6a36294d99a272ac943f959fc6cded748267",
        image = "index.docker.io/sourcegraph/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:e52c0d2dac0380f758cd24122710a5ebd22796067573081207acf869e7a834e0",
        image = "index.docker.io/sourcegraph/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:c9bdd2d6cdd7e75a05de5e9f74484c294bfc3fee28b31bd3bc8027f7b81d9d8c",
        image = "index.docker.io/sourcegraph/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:e8269be7b06ecffe5008f02138d3cb4565637017045a559afde2a2b7028917cc",
        image = "index.docker.io/sourcegraph/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:4644fa597db008b020f8cbc620ed365854dcfd464832b24bc033b2ce067545ec",
        image = "us.gcr.io/sourcegraph-dev/wolfi-grafana",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:afc91aa3285f04c63e651ec90f3f9be7afa2bf9faa21902729844f8b8e4dd9fc",
        image = "index.docker.io/sourcegraph/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:c0d7f9c59f1c772b1a79cc6df22e0ce614128396697c25021db8435f6d42b33f",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:5b92a0286a7ce47918fbd4a4667bb41c6ef0c19491be2f47ee4c66852a6dbeb0",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:222927b690d2cec063de4a508c5ec1456b5692aa66c633484cca5d7743d77431",
        image = "index.docker.io/sourcegraph/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:c3d0c4b2353904cca9242cdc4b6882ed905af1a438f40e0cc2bb62a5d135ff27",
        image = "index.docker.io/sourcegraph/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:0badf5ba5dfffa4bf68a6cc3cb13557fce12d4dfca669b9d62ed729eca831d75",
        image = "index.docker.io/sourcegraph/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:ab536445f71fd8941d7d7300fa519f901da9a4cd8a3b6cfda7bdfaefd37f262f",
        image = "index.docker.io/sourcegraph/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:13f8849c082b705a821542959136485d799cf1cdeee387ae189ee41c4dca30f4",
        image = "index.docker.io/sourcegraph/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:382755edcd88c50bb126e991560bd5e86d74696b1bf14806c90f129afeaa90e4",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:eb91a1da528145946f7ef5c18513ddb1ea14ea4b1f2b6cbf8a612794a5af2d3f",
        image = "index.docker.io/sourcegraph/wolfi-executor-base",
    )

    # ???
    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:da566de7aafa1ab17c2e3a8ebebc425f661667645e2477d247930606e5fbd2d9",
        image = "index.docker.io/sourcegraph/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:1ddc7276bc17c3d71c82c5cd884f9bcd76d3429a87e0a9463198b35e80ba1e33",
        image = "index.docker.io/sourcegraph/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:23b263020d114fae16f141585a4f1f72dcdb961bb7396063b8b746de8460c64e",
        image = "index.docker.io/sourcegraph/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:c52922a7952747dcb777e32f1ab6d82e7c6e25681503dc7675d87f670c0ca7a6",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_prometheus_gcp_base",
        digest = "sha256:bf81591f8c143b9776a1e17efa689f3ee4f3f8880d3d2b1353ab9764cc1f0909",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-gcp-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:e268c4be0ebd84d8ca01d9c6cfdf1c57b070b3f0689bfee75e7ef5e1c6f257ab",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:57260223e0ec9e55071c09c9627dc70a27b2db737d52fa3715fcef78e23a4877",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:90a89460e2debae434613ac6ec8b8846c64b4dfae0e9aab3411befd2fc348ed6",
        image = "index.docker.io/sourcegraph/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:965eaf9c819b6d01182b1f2ad04dfa83bb3aaed63d6d984d1ce880ad86a85936",
        image = "index.docker.io/sourcegraph/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:382755edcd88c50bb126e991560bd5e86d74696b1bf14806c90f129afeaa90e4",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:da4fadc649d9ee0f2a60e3a7f54066894362b9e122062c5070f444d332f4db1c",
        image = "index.docker.io/sourcegraph/wolfi-blobstore-base",
    )

    oci_pull(
        name = "wolfi_qdrant_base",
        digest = "sha256:53513d6f4d32c6e962c690e316e8ad7e3c05bf765c3a7607dce862ddca323293",
        image = "index.docker.io/sourcegraph/wolfi-qdrant-base",
    )
