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
        digest = "sha256:b3dee409328ca79c1f6ad8742461e8b910ee132894eb7baefea787df517a2943",
        image = "index.docker.io/sourcegraph/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:3c4a0611419e8fbbf13f6fd971068064b20fd535d280c411ee04de83d31e96ac",
        image = "index.docker.io/sourcegraph/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:747ab7f58d125a512be6d36a5bb85a7dc2f4da21183904257f85e40b95618141",
        image = "index.docker.io/sourcegraph/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:21071210e80868f891570d657c3915d25eb40da9fabad46f4d71e2fe028b6ba9",
        image = "index.docker.io/sourcegraph/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:5a567927a3a4723c2b5a4eb3bc3dd31fe09761ef3fd2fab72467a5938fc805e2",
        image = "index.docker.io/sourcegraph/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:6f90153789f26885fd708ce695a8f83959216496f55ed9bf9dfe26288a7f1f41",
        image = "us.gcr.io/sourcegraph-dev/wolfi-grafana",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:87130076b52c42e50ea916213817ce186d623f224a6545f833bbe38e0079525a",
        image = "index.docker.io/sourcegraph/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:15252573824c854e0a43ae5f680c3593c4fe1fc3c8c7115a6cbd5a8f1baf61ec",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:f4558bf1ec2523f219d33e77fd56ef47a99b92fbed26ca18c3a3fec07181f5cf",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:09217f91aa350eba6731d9923b2eb3f24bb53f935ee240963523ef5a05f1dc58",
        image = "index.docker.io/sourcegraph/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:34d502280157e91de9f247b37a391875a2a6c42652a6605ce3c0d26ebcd16aee",
        image = "index.docker.io/sourcegraph/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:11f38cf70a61e949cda1cffead24b2a64f8bc8977cd4ac7acad4c15d5a4568b8",
        image = "index.docker.io/sourcegraph/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:64c23ec428ab445c9a505792b794294c053266807563e09d6bae7eb397c9dee5",
        image = "index.docker.io/sourcegraph/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:b777629cf2a83019dddd6d0def0bbc383ab87850b401c92dcc74e1dcb0fa5c38",
        image = "index.docker.io/sourcegraph/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:7afdbbcf70db906e68359162e28658f03165db203fd86f7c7b1582620b2e25d0",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:64672cefaf885e1c890a403865fdddb9fb4aaaecfb284052b9a961d20880466c",
        image = "index.docker.io/sourcegraph/wolfi-executor-base",
    )

    # ???
    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:30c53c672e98942d1d8e7fa600d9c41d341695b1542d3b6f71f9c4507caf775d",
        image = "index.docker.io/sourcegraph/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:2ecc525f015e44f8a4ce3c6ec349d627ee1ad9bbb468c750261cca52515a11b1",
        image = "index.docker.io/sourcegraph/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:65aa30378b45fc999477577af4cdf856c96f6aaf53d72f08a092ee3782ceb3f5",
        image = "index.docker.io/sourcegraph/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:babb794eab89b652a204bfe468f3a0baf72f2db1381ddee0a54a3dfc6496d5c6",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_prometheus_gcp_base",
        digest = "sha256:621c8da9dae5e78eab77b94a3b4c0df536238eb1231957e7ffca4110701445c0",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-gcp-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:d809dd3ab4fd2a0d23dc1fede4e424050b2b090d860be855094058c18977a196",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:9f26754c990a9bc0814c8b15807478b3c22051bb50dfae24521467bb62a96d3b",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:0cc07f7624ca77a236970a99dd3eabb45c9c8256f42dff5b4a427bed84215b7d",
        image = "index.docker.io/sourcegraph/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:60717f967c2f44115589f54a36a5d0bf2dbd2f42bcde69778912c72a674bff06",
        image = "index.docker.io/sourcegraph/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:7afdbbcf70db906e68359162e28658f03165db203fd86f7c7b1582620b2e25d0",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:c46af5725437e745edc1ca138a126fa0d98351eef61f952d63ffc241b5df1bd5",
        image = "index.docker.io/sourcegraph/wolfi-blobstore-base",
    )

    oci_pull(
        name = "wolfi_qdrant_base",
        digest = "sha256:bad099febc367d12ce16955469be3f2c207cf6c86cc73c1cf853e4518e2c4591",
        image = "index.docker.io/sourcegraph/wolfi-qdrant-base",
    )
