FROM golang:1.14.6-alpine@sha256:70d49538b8f7acd5b71e84f81bebf7667f25017308324a91f312a9830a618f3d AS go
FROM sourcegraph/alpine:3.10@sha256:4d05cd5669726fc38823e92320659a6d1ef7879e62268adec5df658a0bacf65c AS lsif-go

ENV LSIF_GO_VERSION=0.8.0

RUN curl -L "https://github.com/sourcegraph/lsif-go/releases/download/v${LSIF_GO_VERSION}/lsif-go_${LSIF_GO_VERSION}_linux_amd64.tar.gz" -o lsif-go.tar.gz && \
    tar -C /usr/local/bin -zvxf lsif-go.tar.gz lsif-go && \
    rm lsif-go.tar.gz

FROM sourcegraph/alpine:3.10@sha256:4d05cd5669726fc38823e92320659a6d1ef7879e62268adec5df658a0bacf65c

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

# hadolint ignore=DL3018
RUN apk update && apk add --no-cache \
    git \
    tini

# Steal latest go from canned build
COPY --from=go /usr/local/go/ /usr/local/go/
COPY --from=lsif-go /usr/local/bin/lsif-go /usr/local/bin/lsif-go

ENV GOROOT=/usr/local/go \
    GOPATH=/go \
    PATH=/go/bin:/usr/local/go/bin/:$PATH

# TODO(efritz) - mount ssd cache
# TODO(efritz) - periodically run `go clean -modcache`
RUN mkdir -p "${GOPATH}/bin" "${GOPATH}/pkg" "${GOPATH}/src" && \
    chown -R sourcegraph:sourcegraph "${GOPATH}"

USER sourcegraph
EXPOSE 3189
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/precise-code-intel-indexer"]
COPY precise-code-intel-indexer /usr/local/bin/
