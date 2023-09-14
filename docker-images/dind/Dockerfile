FROM docker:23.0.6-dind@sha256:eb9f1d80fbe98f6343fd432d8f89db19d5de324096498e67886c4a984f6ac670

RUN rm -f /usr/libexec/docker/cli-plugins/docker-compose && \
    rm -f /usr/libexec/docker/cli-plugins/docker-buildx

RUN apk update && apk add --no-cache 'e2fsprogs>=1.46.6-r0' 'libcom_err>=1.46.6-r0' 'e2fsprogs-libs>=1.46.6-r0' 'libcrypto3>=3.0.8-r3' 'libssl3>=3.0.8-r3' 'openssl>=3.0.8-r3'

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://docs.sourcegraph.com/
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

VOLUME /var/lib/docker
EXPOSE 2375 2376

ENTRYPOINT ["dockerd-entrypoint.sh"]
CMD []
