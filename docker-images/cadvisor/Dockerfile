FROM gcr.io/cadvisor/cadvisor@sha256:9360d7421c5e9b646ea13e5ced3f8e6da80017b0144733a04b7401dd8c01a5cb
LABEL com.sourcegraph.cadvisor.version=v0.45.0

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://docs.sourcegraph.com/

# hadolint ignore=SC2261
RUN apk add --upgrade --no-cache apk-tools>=2.10.8-r0 krb5-libs>=1.18.4-r0 \
    libtirpc>=1.3.2-r1 \
    busybox \
    wget

# Reflects cAdvisor Dockerfile at https://github.com/google/cadvisor/blob/v0.39.2/deploy/Dockerfile
# alongside additional Sourcegraph defaults.
ENTRYPOINT ["/usr/bin/cadvisor", "-logtostderr", \
    # sourcegraph cAdvisor custom port
    "-port=48080", \
    # only enable certain metrics, based on kubelet master
    "-disable_metrics=percpu,hugetlb,sched,tcp,udp,advtcp", \
    # other kubelet defaults
    # see https://sourcegraph.com/github.com/google/cadvisor@v0.39.2/-/blob/deploy/kubernetes/overlays/examples/cadvisor-args.yaml
    "-housekeeping_interval=10s", \
    "-max_housekeeping_interval=15s", \
    "-event_storage_event_limit=default=0", \
    "-event_storage_age_limit=default=0"]
