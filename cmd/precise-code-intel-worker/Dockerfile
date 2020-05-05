FROM alpine:3.10@sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a as libsqlite3-pcre

COPY libsqlite3-pcre-install-alpine.sh /libsqlite3-pcre-install-alpine.sh
RUN /libsqlite3-pcre-install-alpine.sh

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
    tini

# hadolint ignore=DL3022
COPY --from=libsqlite3-pcre /sqlite3-pcre/pcre.so /libsqlite3-pcre.so
ENV LIBSQLITE3_PCRE /libsqlite3-pcre.so
# hadolint ignore=DL3018
RUN apk --no-cache add pcre-dev

USER sourcegraph

COPY ./precise-code-intel-worker /usr/local/bin/precise-code-intel-worker

EXPOSE 3188
ENV GO111MODULES=on LANG=en_US.utf8 LOG_LEVEL=debug
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/precise-code-intel-worker"]
