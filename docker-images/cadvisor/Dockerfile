FROM gcr.io/google-containers/cadvisor:v0.36.0@sha256:16bc6858dc5b7063c7d89153ad6544370eb79cb27a1b8d571f31b98673f7a324
LABEL com.sourcegraph.cadvisor.version=v0.35.0

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://docs.sourcegraph.com/

# Reflects cAdvisor Dockerfile at https://github.com/google/cadvisor/blob/master/deploy/Dockerfile
# alongside additional Sourcegraph defaults.
ENTRYPOINT ["/usr/bin/cadvisor", "-logtostderr", \
    # sourcegraph cAdvisor custom port
    "-port=48080", \
    # only enable certain metrics, based on kubelet master
    "-disable_metrics=percpu,sched,tcp,udp", \
    # other kubelet defaults (v0.35.0)
    # see https://sourcegraph.com/github.com/google/cadvisor@v0.35.0/-/blob/deploy/kubernetes/overlays/examples/cadvisor-args.yaml
    "-housekeeping_interval=10s", \
    "-max_housekeeping_interval=15s", \
    "-event_storage_event_limit=default=0", \
    "-event_storage_age_limit=default=0"]
