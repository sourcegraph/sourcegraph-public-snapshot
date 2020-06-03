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
    tini musl-dev go git

# TODO(efritz) - mount ssd cache
# TODO(efritz) - periodically run `go clean -modcache`

ENV GOROOT=/usr/lib/go GOPATH=/go PATH=/go/bin:$PATH
RUN mkdir -p ${GOPATH}/bin ${GOPATH}/pkg ${GOPATH}/src chown -R sourcegraph:sourcegraph ${GOPATH}

RUN mkdir temp && \
    curl -L https://github.com/sourcegraph/lsif-go/releases/download/v0.8.0/lsif-go_0.8.0_linux_amd64.tar.gz -o lsif-go.tar.gz && \
    tar -C temp -zvxf lsif-go.tar.gz && \
    mv temp/lsif-go /usr/local/bin && \
    rm -rf temp lsif-go.tar.gz

USER sourcegraph
EXPOSE 3189
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/precise-code-intel-indexer"]
COPY precise-code-intel-indexer /usr/local/bin/
