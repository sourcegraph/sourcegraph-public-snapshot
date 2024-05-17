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
        digest = "sha256:ac20eb51c7cc403323d533abf0678afaa893d6a3bff3d97749ab68f5315ba8ab",
        image = "index.docker.io/sourcegraph/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:9ac3c8d78726700c6c499967e7225785d4808b8b0a865f56e15de802c49d9696",
        image = "index.docker.io/sourcegraph/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:118a56f576a0ca0288d2a138756240d472af4d2ce4118258a3c5581e1baa8990",
        image = "index.docker.io/sourcegraph/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:7a83f2f345620b0cbd6062f464ae1ffe587a9662dc3d291f370371299cd7ae63",
        image = "index.docker.io/sourcegraph/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:a8aebb67aeb50423804388219657029c2fafb0efdefa3907d36a0fba580a7619",
        image = "index.docker.io/sourcegraph/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:a493f10eb2a1d76f8923110609735632efb74f90981a787bd3244e51fe603406",
        image = "index.docker.io/sourcegraph/wolfi-grafana-base",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:20c3c69d0bf0a5a2ac25258d649497cf2958f12f0690f96a9cf5c88a8c54589c",
        image = "index.docker.io/sourcegraph/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:a92ade77da3ae1890674307e246287ba9d84420e5e2308a83fcd59c46fe34dfa",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:547a6b76d4bf0f2ce7d2cc3019cf04508a10f4e6ad1dc6369887a8704d41b863",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:bbdbae8501fa344b2dd455ade2903b3cb6c0947e51fbd6e316f1f5ff851979d5",
        image = "index.docker.io/sourcegraph/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:ff177f29c6380348a6b6704eea966ab456c3b142453ff60e03b5a57f49405def",
        image = "index.docker.io/sourcegraph/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:8354d6970d7742793048fcac12e307ac5e096d2e014a3a95d5d07b10b496e592",
        image = "index.docker.io/sourcegraph/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:ff155e2e0be489106415cac7fb15b4ae546221f463732654f79afcde5e4f6f0c",
        image = "index.docker.io/sourcegraph/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:7b6d7bc434683f006667894e1c55fb546815dee699660ea69b8c2710600485d0",
        image = "index.docker.io/sourcegraph/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:c2ead1db5b447fe3d1c422630cd86e59e5e6faf6448b28ff36e193c52173cf1a",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:e685396e1a69a3084422a19863e73e576000b75546641def20d21f684906892e",
        image = "index.docker.io/sourcegraph/wolfi-executor-base",
    )

    # ???
    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:9cf773d806654c221e5776b3a825601f19b643a75a11eda0443742558292aab2",
        image = "index.docker.io/sourcegraph/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:9dc0a526cde08f85c269e26d86de68eb11ea6cd0367d451b7436d205d275f4a8",
        image = "index.docker.io/sourcegraph/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:67e5ac3d28d65ffe1e8b3ab8c8ee1f974d722a4c4f95b8224a0990f7a1a1eef6",
        image = "index.docker.io/sourcegraph/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:791b100132f4524bc5f4c109955e0688c5868822d81398818497a64a8659cbc1",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:504a0b2f9f6c8933784ff24e51db602ab896b689ad77419a261d81852bd713b0",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:1eaafad54d5df9eb20a3da3bbccc6b726db6e16b067e8a652be81f33140d9c41",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:490160056f3c6a152c16a213c6ace754a3874ee88158071a90aeaf9609d0caa2",
        image = "index.docker.io/sourcegraph/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:ed280f1d1b186161971bb8e7083bae8faada1fab21ff25e634c70465f0aec21b",
        image = "index.docker.io/sourcegraph/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:c2ead1db5b447fe3d1c422630cd86e59e5e6faf6448b28ff36e193c52173cf1a",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:22902f226a4b6668ef9314da4ec606c79e89a9e2343f4b05555b114f9ba26193",
        image = "index.docker.io/sourcegraph/wolfi-blobstore-base",
    )

    oci_pull(
        name = "scip-java",
        digest = "sha256:808b063b7376cfc0a4937d89ddc3d4dd9652d10609865fae3f3b34302132737a",
        image = "index.docker.io/sourcegraph/scip-java",
    )

    # The following image digests are from tag 252535_2023-11-28_5.2-82b5f4f5d73f. sg wolfi update-hashes DOES NOT update these digests.
    # To rebuild these legacy images using docker and outside of bazel you can either push a branch to:
    # - docker-images-candidates-notest/<your banch name here>
    # or you can run `sg ci build docker-images-candidates-notest`
    oci_pull(
        name = "legacy_alpine-3.14_base",
        digest = "sha256:581afabd476b4918b14295ae6dd184f4a3783c64bab8bde9ad7b11ea984498a8",
        image = "index.docker.io/sourcegraph/alpine-3.14",
    )

    # https://hub.docker.com/_/docker/tags?name=dind
    # Tag: docker:26.1.3-dind
    oci_pull(
        name = "upstream_dind_base",
        digest = "sha256:ca2dd9db425230285567123ce06dbd3c2aed0eae23d58a1dd5787523a4329eea",
        image = "index.docker.io/library/docker",
    )

    oci_pull(
        name = "legacy_executor-vm_base",
        digest = "sha256:4b23a8bbfa9e1f5c80b167e59c7f0d07e40b4af52494c22da088a1c97925a3e2",
        image = "index.docker.io/sourcegraph/executor-vm",
    )

    oci_pull(
        name = "legacy_codeinsights-db_base",
        digest = "sha256:c2384743265457f816d83358d8fb4810b9aac9f049fd462d1f630174076e0d94",
        image = "index.docker.io/sourcegraph/codeinsights-db",
    )

    oci_pull(
        name = "legacy_codeintel-db_base",
        digest = "sha256:dcc32a6d845356288186f2ced62346cf7e0120977ff1a0d6758f4e11120401f7",
        image = "index.docker.io/sourcegraph/codeintel-db",
    )

    oci_pull(
        name = "legacy_postgres-12-alpine_base",
        digest = "sha256:dcc32a6d845356288186f2ced62346cf7e0120977ff1a0d6758f4e11120401f7",
        image = "index.docker.io/sourcegraph/postgres-12-alpine",
    )
