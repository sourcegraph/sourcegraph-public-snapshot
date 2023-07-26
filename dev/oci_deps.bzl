load("@rules_oci//oci:pull.bzl", "oci_pull")

# Quick script to get the latest tags for each of the base images:
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

def oci_deps():
    oci_pull(
        name = "wolfi_base",
        digest = "sha256:1f41fbd1ad390a4575d6d4107661372e39b875e0a7221bd9f5475a2dcb375f6c",
        image = "us.gcr.io/sourcegraph-dev/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:aae9ff525dbd2c43dcdfd54f74b011d792db7b611dc40ebf9b17387b63f0b52a",
        image = "us.gcr.io/sourcegraph-dev/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:fb906ee43cd89a243aba32efa9c61ccfab50988e2ce220bdcb7da56e59d41026",
        image = "us.gcr.io/sourcegraph-dev/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:c282859aba2866bcb0c504a39e1363df5c29f705890a60f97bf8c112d7eee0b7",
        image = "us.gcr.io/sourcegraph-dev/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:9557a0d248fe7afedba820b32096919fc6b12776ca250ee4bd90d5ab931874ca",
        image = "us.gcr.io/sourcegraph-dev/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:6f90153789f26885fd708ce695a8f83959216496f55ed9bf9dfe26288a7f1f41",
        image = "us.gcr.io/sourcegraph-dev/wolfi-grafana",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:530bc2c44f357bf0e1134b8107591df7f62135336dfb65406f369edb00c39e46",
        image = "us.gcr.io/sourcegraph-dev/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:c6d83b0504bd262dfbcdab4fa7c6a728888e270649554376a4077c21073c3a99",
        image = "us.gcr.io/sourcegraph-dev/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:f72456acb0af84814671f0389ae0aa105b5449a4a991c347009f70aec00ce183",
        image = "us.gcr.io/sourcegraph-dev/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:be1bbdbfd94ea2fad684afd629c5bd8b86e835c2090c92b291366bee2a81973f",
        image = "us.gcr.io/sourcegraph-dev/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:cab23291a0c43c8a47088bdca67ff01ceb1bafe222499b3dda7fa1d45a44a26c",
        image = "us.gcr.io/sourcegraph-dev/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:177a494fc657b95738322d17fa56f0547b948c178abab495f8ee160c849a7458",
        image = "us.gcr.io/sourcegraph-dev/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:ce524267e2d86a4f5f8b05e825a670f6dd01917686ad20f66b7c34bf212b87a6",
        image = "us.gcr.io/sourcegraph-dev/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:b263c76f8d3658afe7305ad93b3dead8f37c51ea9ef619f6e80689dd3f350a55",
        image = "us.gcr.io/sourcegraph-dev/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:8fb353249ed6fa6286568aa12d4dfb7de315292e2455fd790f8387d82c3c4f7c",
        image = "us.gcr.io/sourcegraph-dev/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:aea1f2e51635d20800e19ccf44cb1f18e9988f7a694d51281c0c0cb36a0b5ea9",
        image = "us.gcr.io/sourcegraph-dev/wolfi-executor-base",
    )

    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:44cf4d34ef20e48336446bcbcbc3a26769e4dc7a43e64ca2c0c16c931561faeb",
        image = "us.gcr.io/sourcegraph-dev/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:cd3b91225e3aae076567725205d464f1bd0e5049536179f7839b71b27a42864e",
        image = "us.gcr.io/sourcegraph-dev/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:02d4681a4f9887e46c37745d726e07e9e991735aa0ff3052f6bd0970cf297219",
        image = "us.gcr.io/sourcegraph-dev/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:07dfc9d4654658f3bc783bbec238328cbcbee129edab2b6bede04fcba5b7dd2c",
        image = "us.gcr.io/sourcegraph-dev/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_prometheus_gcp_base",
        digest = "sha256:f5857ce301e98981b1387bc23a59ca7a44487e87a01bcdf010f496e0335e6055",
        image = "us.gcr.io/sourcegraph-dev/wolfi-prometheus-gcp-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:0f521abae263be472d8049904375c0e48f40ed560088a418b33616a9bb6bacdc",
        image = "us.gcr.io/sourcegraph-dev/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:b362b9313df22e69214c30fd660b21bfe2354921a0b915efcb51ffc288c099eb",
        image = "us.gcr.io/sourcegraph-dev/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:3cd6ae2b9c59c230ae0f65a6aa2e8a030361de82d4273ed64f5cf8052c94f197",
        image = "us.gcr.io/sourcegraph-dev/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:9c96f8fb5197ea6a51272cd9d7c7bb479290a2aadd11de950b144882a67392e9",
        image = "us.gcr.io/sourcegraph-dev/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:8fb353249ed6fa6286568aa12d4dfb7de315292e2455fd790f8387d82c3c4f7c",
        image = "us.gcr.io/sourcegraph-dev/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:09239985070a367d1f1d839bed5326af56d9dcf2133b779456b8255f5053e4c6",
        image = "us.gcr.io/sourcegraph-dev/wolfi-blobstore-base",
    )
