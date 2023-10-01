FROM docker:23.0.6-dind@sha256:afa5d51349001a4293bab02d2988290db5eae393a83405b1f458c0dd44e2f3ca

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
