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
        digest = "sha256:147dc685d4a8c0aefbe97c5cc26801fb0db51fe93824fc56b193a7ca57af2eb1",
        image = "index.docker.io/sourcegraph/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:85e6e9a36531c4d69e943ecd75ec9b2a805e10dfc94a590219f5ca9ddc2d0319",
        image = "index.docker.io/sourcegraph/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:fde046db38e63943ba8c1a5132f070260b37943c502cdcc225382931f7389432",
        image = "index.docker.io/sourcegraph/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:380cebee4eb173f4d95932c1d03e12f1fbc6d2da8a6289477a20587f5f6959d4",
        image = "index.docker.io/sourcegraph/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:feb5317360d0460ae6b3673f794126dff928adf0dc234f3833613d1c2ec65977",
        image = "index.docker.io/sourcegraph/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:1989633b60c45edf330f1161f7be37bc75b905b0947607e175bb82db224ce50a",
        image = "us.gcr.io/sourcegraph-dev/wolfi-grafana",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:6a6417f50496c149a117267f0f1d74a23a1122249dbe3d34c8ee0dc91fa18e82",
        image = "index.docker.io/sourcegraph/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:c50585382b90f8144f2231ab55a8af0c9d9f207b43c13c400a1290986f9db32f",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:b343c2c15520382692cb930ab7379d5aa14f74e45e157073daac1c9b4b2f69a5",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:ef217a0dacb607448bc49aabaf491d433184c3dab8c905580656114dbdab87d1",
        image = "index.docker.io/sourcegraph/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:3569863f553320c43f842c0cd2e2750c47ed58501dc55327b731782f0bf01d47",
        image = "index.docker.io/sourcegraph/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:9c1fe6d59d068c2ce6184a4f38b08554c1a140000c16f62af98957750451d026",
        image = "index.docker.io/sourcegraph/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:25836c7194d3510d4d0542c35960d48c1c9b476cf1d78b87e015c2ecfbeb29a7",
        image = "index.docker.io/sourcegraph/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:b2c4daa5495a194710e3ea4e44e3d02ee83ce28248cd1611c978b4faf1971fad",
        image = "index.docker.io/sourcegraph/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:2f2b66726ce252b08ef0211737878b963bbc5425c913317bb1883e90c852bb7c",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:8a2bae2576d60387249966faa1ad0a4c899f68f4489f7561816f0f6ade989598",
        image = "index.docker.io/sourcegraph/wolfi-executor-base",
    )

    # ???
    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:17aed9c66d97aa7ad9bbdea61ae0658d34e50f642bdc4404f94c0eb130a4ae67",
        image = "index.docker.io/sourcegraph/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:200a75bd531508d389a22b68b7dc91fce709961bad89d28ec39ffac4bf38e895",
        image = "index.docker.io/sourcegraph/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:b5a811ac666da7e8a8cd0ccccc721c85e0403ae48ae69d685ae6a4fb0982f061",
        image = "index.docker.io/sourcegraph/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:ca0b0de4206eebf978fb7a1a15b98be6f08f21be7ff530e0e28ba0424830b2a9",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_prometheus_gcp_base",
        digest = "sha256:af2aa1a9ab61b43e0bfb2db2e51eb9eec75ef4365473ff2ce3cfdce3eaf0bde7",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-gcp-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:df4e79a3cc0d5f95500d8639271ac67953ee4cb5cb2fb467a8d34bb34c1f258b",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:359f9d7b0aeaa80ac5258d404a9e754818410912da4098f8e27436edd18d019e",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:a592dc28fc04587c6973eeb8b3c6dc035c604d598cdfefde63d5b36926817171",
        image = "index.docker.io/sourcegraph/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:ee09de497a2351f70904d71cf853ceb539e55944f53d8d15a29bfe14499eb237",
        image = "index.docker.io/sourcegraph/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:2f2b66726ce252b08ef0211737878b963bbc5425c913317bb1883e90c852bb7c",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:a051b85a71f186eb91e75de57e96ddc684e65da13246fb9da827e3af2242f084",
        image = "index.docker.io/sourcegraph/wolfi-blobstore-base",
    )
