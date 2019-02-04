FROM alpine:3.9 AS ctags

# hadolint ignore=DL3003,DL3018,DL4006
RUN apk --no-cache add --virtual build-deps curl jansson-dev \
  libseccomp-dev linux-headers autoconf pkgconfig make automake \
  gcc g++ binutils

ENV CTAGS_URL=https://codeload.github.com/universal-ctags/ctags/tar.gz/7918d19fe358fae9bad1c264c4f5dc2dcde5cece

# hadolint ignore=DL3003
RUN curl -fsSL -o ctags.tar.gz "$CTAGS_URL" && \
  tar -C /tmp -xzf ctags.tar.gz && cd /tmp/ctags-7918d19fe358fae9bad1c264c4f5dc2dcde5cece && \
  ./autogen.sh && LDFLAGS=-static ./configure --program-prefix=universal- --enable-json --enable-seccomp && \
  make -j8 && make install && cd && \
  rm -rf /tmp/ctags-7918d19fe358fae9bad1c264c4f5dc2dcde5cece && \
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
  'bash=4.4.19-r1' 'postgresql-contrib=11.1-r0' 'postgresql=11.1-r0' \
  'redis=3.2.12-r0' bind-tools ca-certificates git@edge \
  mailcap nginx openssh-client su-exec tini

# hadolint ignore=DL3022
COPY --from=sourcegraph/syntect_server:d74791c /syntect_server /usr/local/bin/
COPY --from=ctags /usr/local/bin/universal-* /usr/local/bin/
COPY . /
ENV GO111MODULES=on LANG=en_US.utf8 VERSION=$VERSION
ENTRYPOINT ["/entrypoint.sh", "/sbin/tini", "--", "/usr/local/bin/server"]
