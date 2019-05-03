FROM alpine:3.9@sha256:644fcb1a676b5165371437feaa922943aaf7afcfa8bfee4472f6860aad1ef2a0 AS ctags

# hadolint ignore=DL3003,DL3018,DL4006
RUN apk --no-cache add --virtual build-deps curl jansson-dev \
    libseccomp-dev linux-headers autoconf pkgconfig make automake \
    gcc g++ binutils

ENV CTAGS_VERSION=5b4865cc2d4831db9d638a647ff2f5a0dced2133

# hadolint ignore=DL3003
RUN curl -fsSL -o ctags.tar.gz "https://codeload.github.com/universal-ctags/ctags/tar.gz/$CTAGS_VERSION" && \
    tar -C /tmp -xzf ctags.tar.gz && cd /tmp/ctags-$CTAGS_VERSION && \
    ./autogen.sh && LDFLAGS=-static ./configure --program-prefix=universal- --enable-json --enable-seccomp && \
    make -j8 && make install && cd && \
    rm -rf /tmp/ctags-$CTAGS_VERSION && \
    apk --no-cache --purge del build-deps

# TODO: Make this image use our sourcegraph/alpine:3.9 base image.
FROM alpine:3.9@sha256:644fcb1a676b5165371437feaa922943aaf7afcfa8bfee4472f6860aad1ef2a0

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

RUN echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories && \
    echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories && \
    echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/v3.6/main" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/v3.6/community" >> /etc/apk/repositories

# hadolint ignore=DL3018
RUN apk update && apk add --no-cache \
    # NOTE that the Postgres version we run is different
    # from our *Minimum Supported Version* which alone dictates
    # the features we can depend on. See this link for more information:
    # https://github.com/sourcegraph/sourcegraph/blob/master/doc/dev/postgresql.md#version-requirements
    'bash=4.4.19-r1' 'postgresql-contrib=11.2-r0' 'postgresql=11.2-r0' \
    'redis=3.2.12-r0' bind-tools ca-certificates git@edge \
    mailcap nginx openssh-client su-exec tini

# hadolint ignore=DL3022
COPY --from=sourcegraph/syntect_server:5e1efbb@sha256:6ec136246b302a6c8fc113f087a66d5f9a89a9f5b851e9abb917c8b5e1d8c4b1 /syntect_server /usr/local/bin/
COPY --from=ctags /usr/local/bin/universal-* /usr/local/bin/
# hadolint ignore=DL3022
COPY --from=libsqlite3-pcre /sqlite3-pcre/pcre.so /libsqlite3-pcre.so
ENV LIBSQLITE3_PCRE /libsqlite3-pcre.so
COPY . /

RUN echo "hosts: files dns" > /etc/nsswitch.conf

ENV GO111MODULES=on LANG=en_US.utf8
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/server"]
