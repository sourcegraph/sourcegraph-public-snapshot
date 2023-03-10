FROM docker:23.0.0-dind@sha256:be7c1cb42809b910473a8a1b195736758fc8c10b395001b90968d5f31ad6a40b

RUN rm -f /usr/libexec/docker/cli-plugins/docker-compose && \
    rm -f /usr/libexec/docker/cli-plugins/docker-buildx

RUN apk update && apk add --no-cache 'e2fsprogs>=1.46.6-r0' 'libcom_err>=1.46.6-r0' 'e2fsprogs-libs>=1.46.6-r0' 'libcrypto3>=3.0.8-r0' 'libssl3>=3.0.8-r0'

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
