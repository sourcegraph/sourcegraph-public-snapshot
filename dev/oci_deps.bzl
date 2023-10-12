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
        digest = "sha256:3daf91b2f72d23315e8af6c48ac9b6a9970e124a50bfe171326605ff46906f3b",
        image = "index.docker.io/sourcegraph/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:20802f3e99f1ed40a1c206cf5cef301b63749ae900c207d4ff7d43c149ba7762",
        image = "index.docker.io/sourcegraph/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:6b6812d22fe8598693d8b22e0e1957b6a0a2687cacb3557fc38b49bc8687131f",
        image = "index.docker.io/sourcegraph/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:248bf90055c05eaaf1195e14b61ffe9780784531e9c1ae7a6a1471903d984995",
        image = "index.docker.io/sourcegraph/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:026e539acdf4c7c5d3660d54e2058447d30029819913e27d94e3d99419311fc8",
        image = "index.docker.io/sourcegraph/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:4644fa597db008b020f8cbc620ed365854dcfd464832b24bc033b2ce067545ec",
        image = "us.gcr.io/sourcegraph-dev/wolfi-grafana",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:e395440b783b31f1589e00e3b131bbce8d9ae24a88d36ee48c7725bc80a8b20f",
        image = "index.docker.io/sourcegraph/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:a92722ab54b485380d4a932451e777c9b3893be64f3c07008d524136fc4e0e3b",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:ad4aa059111176b16457b3c5eb7381b769c48e7ce0139308283a64d8827ee72f",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:b4f5eb9445efc17c9a11a4e56b701fbe6a57bac9cb5791b6d809f3c08e485bfd",
        image = "index.docker.io/sourcegraph/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:dbc89dc7f4b3828f9e90cf23db32ecd27243269ff36da60b1a2f5dab644a845a",
        image = "index.docker.io/sourcegraph/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:1dc9342231803a603662b05c9f91437dfc2b41124306f32b9c6ba117a5c142a0",
        image = "index.docker.io/sourcegraph/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:98bc3f1d5154b86a03dcca6ec503ae6c1a580b5011479fdce24628db551c2938",
        image = "index.docker.io/sourcegraph/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:d77a8155dc0ef1a04a96d54883f0c779e01d60898b4fce6779f00b23d183955c",
        image = "index.docker.io/sourcegraph/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:0bf7e3a4ec792a0507b0dad2ae30000e86b21cd54230bc1f93febeeae071edaa",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:6e01b23cbf988c08b5f80719004f190481213ac89d2f9fec5a35f7bf5b74b050",
        image = "index.docker.io/sourcegraph/wolfi-executor-base",
    )

    # ???
    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:8e79a3785bc1d7a5ab33032ccbb086e557ffc1c81de645404abb5fa815df3dd7",
        image = "index.docker.io/sourcegraph/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:4e1f8f71749c6e5b0ee672c866812af4d4ccfb2ad7e11ac7f4e5a60967425bca",
        image = "index.docker.io/sourcegraph/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:803a7b37e63e5749361f74b213de31147ac92de73463a78c61f230698ce3aa0d",
        image = "index.docker.io/sourcegraph/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:07fd05cb46f46643c326d7bfe9054ba44b1940815465083ac64dec55b65d9775",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_prometheus_gcp_base",
        digest = "sha256:f0f7299e030dc8de0268a10fcfcf415bf895f9a3e777dd63d3e7f4008c7262b5",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-gcp-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:4ca4deb370690c9ca26e9eccb4e71550cb47c7404a3107a3e7e4c9fa9aa37bf9",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:8205a0c561e23acb8b7552ce5a81699650739c00aa133ce6294540b6beb7e57b",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:45554509dd6013763503103541465ebc6a5760c02b523f52c8e08bafb8741d9a",
        image = "index.docker.io/sourcegraph/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:c8a6309a8f1fe62311f8bef3d2971441eb411c54c566fdcea217e23792dfc772",
        image = "index.docker.io/sourcegraph/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:0bf7e3a4ec792a0507b0dad2ae30000e86b21cd54230bc1f93febeeae071edaa",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:e1710fb154982b67e14b1db8bc324e5ec411837539c40075be4855aa81c9f8f1",
        image = "index.docker.io/sourcegraph/wolfi-blobstore-base",
    )

    oci_pull(
        name = "wolfi_qdrant_base",
        digest = "sha256:80617026086a76f90834bcab08a8986850d2b1921c6e4fb15e87b236c402aa15",
        image = "index.docker.io/sourcegraph/wolfi-qdrant-base",
    )
