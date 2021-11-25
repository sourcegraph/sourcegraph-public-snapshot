FROM sourcegraph/alpine-3.14:117816_2021-11-25_6eb4943@sha256:06d0ecb92132c7c40ce6476b2e72d633b55b1076a49411526985fa08cdfc16ab

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

# hadolint ignore=DL3018
RUN apk update && apk add --no-cache \
    tini

USER sourcegraph
EXPOSE 3189
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/worker"]
COPY worker /usr/local/bin/
