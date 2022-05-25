FROM gcr.io/cadvisor/cadvisor@sha256:8938726fe00fd7a3889f7c4fb50a54b728f1d02fb5f6cbdbea604824ad11ff3f
LABEL com.sourcegraph.cadvisor.version=v0.44.0

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
