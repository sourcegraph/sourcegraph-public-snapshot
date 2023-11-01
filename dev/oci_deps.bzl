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
        digest = "sha256:5b41eef0aa853b3cdf1b91f8955506caa3984af88788eceeb4f47f3e7b45ce99",
        image = "index.docker.io/sourcegraph/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:42ae3d0a4fe6e4df4c993fd7016d461cec64f3d1c599111158683557c79d136c",
        image = "index.docker.io/sourcegraph/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:8407df261fa19d1bd113b43746a3c4aebc493e7a6cdea4a29043fd4a4ea54845",
        image = "index.docker.io/sourcegraph/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:f7e9b0c94230dd6329ae31a8be3f1f5b58b51ffecbd0342e448556bea985490c",
        image = "index.docker.io/sourcegraph/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:b62969dc5be16dd8bf7040133bae1706967cbff1573968e980bc819feb53a4e7",
        image = "index.docker.io/sourcegraph/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:07a31e9c49300872deeb162910207611261475634bd7bc92e4dcf9a285c98f16",
        image = "us.gcr.io/sourcegraph-dev/wolfi-grafana",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:5749193cdb45ff965443138ce79a76656c30132d7f3d6eaf0392bb1d863a2b37",
        image = "index.docker.io/sourcegraph/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:50c2c05515de819bd28e487f77b31cf1354639ddf8827e35b9dfb4df4a365c9f",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:c08cfc258c596ef3d1cd1da1d6d1f812d374537dc84e26ed022c1613c2843d6f",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:8f4da841953499bc557f3b419b9689bec35a6fc50e386160585433c24e8bae19",
        image = "index.docker.io/sourcegraph/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:064f730d051044b686a1561234a740f7ba2da7adb7d0ba3a8ee3e1ce31ea207f",
        image = "index.docker.io/sourcegraph/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:17fce6bf849230494f82cd76bd8b0ac8ff69f9f92fff394b396d7d458abc87ba",
        image = "index.docker.io/sourcegraph/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:cfc20ad745e364e41fcf7edf54113479cabd0a60017d2c8b6be7840f6490e5eb",
        image = "index.docker.io/sourcegraph/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:09c1516656686a01968da5b41963bd60c07ba3eb6e2503840b6d835d3cd31917",
        image = "index.docker.io/sourcegraph/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:1cdef57c14f194d156425c040f5f02845888bf470bc9dd19613e32f4959040bb",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:ab28fbdb0c62f9e82cc0a69ceaf9683401205bb1e235e4761ea7a383cc214554",
        image = "index.docker.io/sourcegraph/wolfi-executor-base",
    )

    # ???
    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:fbf96f74464d64a52a43322d00c038389accac967dbb410277a6bb83105f1ef0",
        image = "index.docker.io/sourcegraph/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:1c6235d751e2ef9fc73939998f0f7a1e1751292247c28d33abe9531efdadde9f",
        image = "index.docker.io/sourcegraph/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:f4b7916c4ddc3339a77e088ab5efcbce0092f3938742250992cf4a5845ff0042",
        image = "index.docker.io/sourcegraph/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:c8b12e6d46290ff86f6b2bf0423ede40b37adc7a9eac9d2bad8ca623edbee771",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_prometheus_gcp_base",
        digest = "sha256:11553d06f8e8e50b7587bee1194c7d2d594113a5cb0dfde8cc239a9f3f65b1d0",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-gcp-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:02386691fdb41cf9a02c877545f2baf5b8d1b6e8c55908832373c8b213a88c42",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:4c3f31fd3bef507042939a37fd1d8cdf8e837e66f57697ff9d57703e3e2533d0",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:6ec6a5a76e1a87c5a293c4b2c379263c2774093b48ffc0b834114494e590cb67",
        image = "index.docker.io/sourcegraph/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:48c6ded19a8a535cc7242fad4968ac18e7be71cd03b8722d39524d8a457c54ad",
        image = "index.docker.io/sourcegraph/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:1cdef57c14f194d156425c040f5f02845888bf470bc9dd19613e32f4959040bb",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:70d821a8474ab472045ca6bc06227ffb3b8540e2c94d8684062e06d5b7af2e89",
        image = "index.docker.io/sourcegraph/wolfi-blobstore-base",
    )

    oci_pull(
        name = "wolfi_qdrant_base",
        digest = "sha256:0ffa70add77182bf112de1c27ff3568e1de511ccc763dbc2f7530075f1be7e94",
        image = "index.docker.io/sourcegraph/wolfi-qdrant-base",
    )
