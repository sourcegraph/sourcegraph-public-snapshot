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
        digest = "sha256:a31afa210aaf90f08bc757cc4de12f84de2d311e51988280d0e429f36ad868ec",
        image = "index.docker.io/sourcegraph/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:ed091dd8a0b408d9a9cccdb3e1b33ee02da6991dc71344873cb7c7154b0d84b7",
        image = "index.docker.io/sourcegraph/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:2f6fed42dcb55bd2676a27cbcc213f10024139b22b79305c0cbd0c6e0d551136",
        image = "index.docker.io/sourcegraph/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:8a887a147e18cf13bc39ee10965e9fa9e26dd68bd854e7bf438bc9b2bdaac133",
        image = "index.docker.io/sourcegraph/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:617db82d904658d803ed3b133a9a4583991ba9fd14b2c35c255e14c407ff811c",
        image = "index.docker.io/sourcegraph/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:1f56cfd90b7412e9784434c891427142c6e3238d6fb8602bd5e04333a5214daf",
        image = "index.docker.io/sourcegraph/wolfi-grafana-base",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:226a6f06cad1e96b22446c00340569c64e1047f48f5ac8ce25a900bcc666a255",
        image = "index.docker.io/sourcegraph/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:f1236a7659232568caae67cc3de42e415a53f0a4291fe0c765a3dc5280bd11ff",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:0a08f1d419fbcb2df2438fe19ddc8ef13fe89178f4a7dcdfc4bda270bf3c9140",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:081b93e4397b83c303c275e10894773ed55b07eee0daa5d608e5f234d4407ee3",
        image = "index.docker.io/sourcegraph/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:e116b4e47b4cc8b0de48dae215593c45ff737f87479ff17e0f90a79adf8c5873",
        image = "index.docker.io/sourcegraph/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:ceda5c21f8461d7351866214c11b65c2c695d4c458808b93e2be6260729cc5cb",
        image = "index.docker.io/sourcegraph/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:4fbe187a00b9af4e2a88af49b12a601769a283e19f61c4a51fd58afc2e299b05",
        image = "index.docker.io/sourcegraph/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:01dc7ff7adeb7557c903a3eec1fce531e5cb0c01dd66eb397150756f8af74dad",
        image = "index.docker.io/sourcegraph/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:9f77f0b0cc95fbb5d8083ab9ea6fcd54c1e936b1ae9c52887c0b4e8b6d34b763",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:b979e0cfb730262538e3e6845644d80b99a622476a663433712f6e74f8c0e67d",
        image = "index.docker.io/sourcegraph/wolfi-executor-base",
    )

    # ???
    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:a5c08b69b1083ba22d50bfb94c7c22b4222948b2d664c39936502b3005c3426f",
        image = "index.docker.io/sourcegraph/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:fc6fc2563d33fb0c65f17712ad4390565ab8656d3b5b0b2558227fba65d0d7d6",
        image = "index.docker.io/sourcegraph/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:6030276d0a702cd386ec86907ce6905da21edbf137e73a682fefa31dc7795005",
        image = "index.docker.io/sourcegraph/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:419c06e9cd4183cdb5a85be2aaac436d3d24b01d9fe1670b6ab5c76ebc788f3a",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:c85c0b034a7c983e000c46fd11aa175256078f7e9531b41ab2ba70635571db76",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:d007565b59a65389e32b79c371e6c558acf3283cdd856c246d731f49886dddc9",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:f7873465e2a383f2973746d339bd04b2e933e5fe8d4bf73baecd12dbf34ed0da",
        image = "index.docker.io/sourcegraph/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:e8e15d5301e3128aabfba930c3a9f693bb3dbf0ef3e9cf6e790c84cc48d2ea21",
        image = "index.docker.io/sourcegraph/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:9f77f0b0cc95fbb5d8083ab9ea6fcd54c1e936b1ae9c52887c0b4e8b6d34b763",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:732027e3a004a60591b1ead52d63ef8ec24768079c0ad6f3d0b1f782745b7d63",
        image = "index.docker.io/sourcegraph/wolfi-blobstore-base",
    )

    oci_pull(
        name = "wolfi_qdrant_base",
        digest = "sha256:9a7f31255d025fc762530360d4f056af98ac5ae67c2258ff8eae8afeace60587",
        image = "index.docker.io/sourcegraph/wolfi-qdrant-base",
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

    oci_pull(
        name = "legacy_dind_base",
        digest = "sha256:0893c2e6103cde39b609efea0ebd6423c7af8dafdf19d613debbc12b05fefd54",
        image = "index.docker.io/sourcegraph/dind",
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
