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
        digest = "sha256:dc44581c10496b9a85ee63b91ec679f3942b9cef256762956b860db8001a71b4",
        image = "index.docker.io/sourcegraph/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:5b40d5db451baa54d4ea26da900337c0524b871f2dbe0871ab7022be6dccd0dc",
        image = "index.docker.io/sourcegraph/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:ab3d9d69e5d216917a59f68d4151e0ab7bb461f55844bfca724d694a48089513",
        image = "index.docker.io/sourcegraph/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:7399778f564fcaa1f41eddc4dc370d3fa3fcebc0062b2ff9aace618728a96d86",
        image = "index.docker.io/sourcegraph/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:2be0320b8fb025f0786c70d485df98c3e69f86569ae661a665e72553056eb16b",
        image = "index.docker.io/sourcegraph/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:e9dcc49846ad3013869b2f070dd51eed109e740a2e330f5286a7a9a82b1f4ae5",
        image = "index.docker.io/sourcegraph/wolfi-grafana-base",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:5b4ca7bb4598c0fe1cc01128527dac093878612ecc6059dcee53480124f6b588",
        image = "index.docker.io/sourcegraph/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:73f450e004e076f41a0715e4210f906c27a9c1beb25ed86a6c616298352e268c",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:52e5092360f40b774931fced86983118dbdac382025fac437dc1bb936305a38b",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:53951db2752e085b4fb8791e1cd46d521d68f2c160661205f803a377e54ed3e9",
        image = "index.docker.io/sourcegraph/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:4a1b9c391e97994e16cb19fba77fcc05b1af5c2928cf094712b3a182dac6bdd9",
        image = "index.docker.io/sourcegraph/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:aad68d003c7fa26cd22144116332beef0f1e55e48b17d9cb46fd98fbf748c0aa",
        image = "index.docker.io/sourcegraph/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:efc7a8f70279f85eed88a862d77d0fcc8f87582f2df15c774b602cddfa61b305",
        image = "index.docker.io/sourcegraph/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:ce106fddf98d4cd0e5793b11fa207171208917a5629863b799071b5306e31991",
        image = "index.docker.io/sourcegraph/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:132757381aefc1f7bff267a09f64a32fe2d57d97a42b9f692d9787bc8b9b1940",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:25a9928c00b33a49257e92a11597bebe09a1ff8bc1adedf8bc903985adfca72c",
        image = "index.docker.io/sourcegraph/wolfi-executor-base",
    )

    # ???
    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:4f5a87784f603433f4b7e515fde56188c8c7413939fe3bfef7bd9f2e9073b626",
        image = "index.docker.io/sourcegraph/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:badfaf46388e8a0b5ac96389bd6d364a5f6d4bca56ba62dfe1b8e977910ee804",
        image = "index.docker.io/sourcegraph/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:69675f48a8b0562c34b08ce8f1a6d7607466d3d69635c383e0b6e6552ea3f0f6",
        image = "index.docker.io/sourcegraph/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:89051f8231b31e12c3ae9ee3db637823e0490025b0aac51d5d5c8c7d68e63f22",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_prometheus_gcp_base",
        digest = "sha256:65ee86af8fd1d8a4d81785ae02cce436e5dec42b883228e091b6850bc87aae25",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-gcp-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:47e61bc8b4a5f73e5f344f13cf908807d98125c3bb0668bca6a57d81e9cd3e42",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:418be77a09486a48dd53c843909f16b2ad06b5a848f95c7064f250da0853b1bd",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:155043d91589010a5a07f78d2861931e69fb2ae6977be7bb18ed318512b12e79",
        image = "index.docker.io/sourcegraph/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:4a81c147937176777e9fbdbf4066bdbe3eab5dd432f74fa3199f58917f7f1b42",
        image = "index.docker.io/sourcegraph/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:132757381aefc1f7bff267a09f64a32fe2d57d97a42b9f692d9787bc8b9b1940",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:50a4535656af20af740d13da7e7bc4abfb2c35b80fef2322f0fd79d688e5a820",
        image = "index.docker.io/sourcegraph/wolfi-blobstore-base",
    )

    oci_pull(
        name = "wolfi_qdrant_base",
        digest = "sha256:c1dc38553bca3097c92be1aa66ec3f198593f16ccbbbe935504316c1218f92bb",
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
