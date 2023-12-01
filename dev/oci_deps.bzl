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
        digest = "sha256:02125b1618c5ea5be1ca880a21101f9bc28d8548bc7706b1386fe8f66e2772ba",
        image = "index.docker.io/sourcegraph/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:e8e2110ae3e3099868465a81b7f19409db3352327f5bced9f3a263017fc70fa9",
        image = "index.docker.io/sourcegraph/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:e7c116f5d2e41a0693dcbecca5e132a4fcbb5bbe6e011b14c1eb661cdaeaf662",
        image = "index.docker.io/sourcegraph/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:43e64fa0fd3e2eb43f4c5e5653f1a27517349703ea3048a80e99082973707202",
        image = "us.gcr.io/sourcegraph-dev/wolfi-server-base",  # use stable p4-fusion-binary
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:b7a8918dba20363f50b0fed1cd4aea4e04522788bc92ad07fedf065554085901",
        image = "us.gcr.io/sourcegraph-dev/wolfi-gitserver-base",  # use stable p4-fusion-binary
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:461ef0c0233c20716398a0ad1a38dc2df955e26ec091400f64a18fa7de9d2621",
        image = "us.gcr.io/sourcegraph-dev/wolfi-grafana",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:4ed9ecc3c4dfe0c344ff7a733e020b957db6f6afa191f2707bcceaaac21bd349",
        image = "index.docker.io/sourcegraph/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:a37be511cf350c588b065305a2e64dbce235cd1b3039bdc5cf3fa6f4821ff5bf",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:a55ed55e4c58365c8316c5cf3bcfb458eadf1d4d230e0c39ab881ffae7074af3",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:7a8bd23e6169def08b449ee251fa0570e5248c665a7ddc99bf68e3f868251b25",
        image = "index.docker.io/sourcegraph/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:9f7c43cf12e2238869a17dfb91320a8a21b446948ca12e10d9130b00957f569d",
        image = "index.docker.io/sourcegraph/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:325acc79b8705c12b64b2395c2d256d7923d2fd5ba075f2b77cea935d2d9778b",
        image = "index.docker.io/sourcegraph/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:0a78eb4d55e9081ef96e81a8d18ae42cd23cfba23ded03079852b76ec3507b64",
        image = "index.docker.io/sourcegraph/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:8bc5c3a75072b817bca552eaa1f864fa420c2dcbf6fa2246fd96d184a28a50a2",
        image = "index.docker.io/sourcegraph/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:705c6cfd121dc08c3d12dd15bac9b4cca1184fcf85b16a84a191be2db2310456",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:1426896be04ac94901fb19180cba4c2956797b478ddd21a626ed6f4627bbf7de",
        image = "index.docker.io/sourcegraph/wolfi-executor-base",
    )

    # ???
    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:64351d526a3e69b34f30ee9f00933c3f52b080c7d43b196ef90710e1e06d74ac",
        image = "index.docker.io/sourcegraph/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:5cd8c62d55afebd62beca290abd035b9e704cd6cdf426fd7799c56b96417f29c",
        image = "index.docker.io/sourcegraph/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:e0c85de735bbd2d7f217bc9f7d1fb71d329d632a3f34533c6090b9da4ce014a4",
        image = "index.docker.io/sourcegraph/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:782bf0b473e1125aafbf38ccec801d966c2ee95e81b2bbc48575d34f994f3eb6",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_prometheus_gcp_base",
        digest = "sha256:af3a02d81908cd1b82f76e88cc5197c241c94f6a034417488fa7eecb4d946141",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-gcp-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:1414b44bbf805c9c853c011d79f863c605ba279d84c38c3a2f0d236934a24955",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:7a906a558341e629f3725e988798a09a396ff5cec4881469d1b007195001dd9e",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:5848af6c50ea346d11ce19b65b55e9f00246d984ff61caf84bd2ffb0103be9dc",
        image = "index.docker.io/sourcegraph/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:826148b4b82f0f63863493384bf30e80b253ade8a4f5e948e4eb5be0b441c44a",
        image = "index.docker.io/sourcegraph/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:705c6cfd121dc08c3d12dd15bac9b4cca1184fcf85b16a84a191be2db2310456",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:26ddd2937a01de167b4e70161cc1be908345c9afcbcc1d05fea699ac818d8b01",
        image = "index.docker.io/sourcegraph/wolfi-blobstore-base",
    )

    oci_pull(
        name = "wolfi_qdrant_base",
        digest = "sha256:988b95f8219ddfdaa7e5dcedec0eceb7d425fa0f2b7179d026f1e887e4f19a7f",
        image = "index.docker.io/sourcegraph/wolfi-qdrant-base",
    )
