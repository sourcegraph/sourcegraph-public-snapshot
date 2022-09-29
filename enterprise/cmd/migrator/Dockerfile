FROM sourcegraph/alpine-3.14:174825_2022-09-28_291dca5d9725@sha256:0bfed82a206a1f313b7276521375dc4d27caa45c324f99e7da0453cecbd31980

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

RUN apk update && apk add --no-cache \
    tini

USER sourcegraph
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/migrator"]
COPY migrator /usr/local/bin/
