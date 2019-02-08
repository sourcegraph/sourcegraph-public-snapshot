FROM alpine:3.9 AS ctags

# hadolint ignore=DL3003,DL3018,DL4006
RUN apk --no-cache add --virtual build-deps curl jansson-dev \
  libseccomp-dev linux-headers autoconf pkgconfig make automake \
  gcc g++ binutils

ENV CTAGS_VERSION=08a69b5a030cf56c079e3006238b992b34c9cc51

# hadolint ignore=DL3003
RUN curl -fsSL -o ctags.tar.gz "https://codeload.github.com/universal-ctags/ctags/tar.gz/$CTAGS_VERSION" && \
  tar -C /tmp -xzf ctags.tar.gz && cd /tmp/ctags-$CTAGS_VERSION && \
  ./autogen.sh && LDFLAGS=-static ./configure --program-prefix=universal- --enable-json --enable-seccomp && \
  make -j8 && make install && cd && \
  rm -rf /tmp/ctags-$CTAGS_VERSION && \
  apk --no-cache --purge del build-deps

FROM alpine:3.9
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
  'bash=4.4.19-r1' 'postgresql-contrib=11.1-r0' 'postgresql=11.1-r0' \
  'redis=3.2.12-r0' bind-tools ca-certificates git@edge \
  mailcap nginx openssh-client su-exec tini

# hadolint ignore=DL3022
COPY --from=sourcegraph/syntect_server:d74791c /syntect_server /usr/local/bin/
COPY --from=ctags /usr/local/bin/universal-* /usr/local/bin/
COPY . /

ENV GO111MODULES=on LANG=en_US.utf8
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/server"]
