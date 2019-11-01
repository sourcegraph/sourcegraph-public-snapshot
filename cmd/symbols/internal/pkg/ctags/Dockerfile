FROM alpine:3.9@sha256:bf1684a6e3676389ec861c602e97f27b03f14178e5bc3f70dce198f9f160cce9 AS ctags

ENV CTAGS_VERSION=03f933a96d3ef87adbf9d167462d45ce69577edb

# hadolint ignore=DL3003,DL3018,DL4006
RUN apk --no-cache add --virtual build-deps curl jansson-dev libseccomp-dev linux-headers autoconf pkgconfig make automake gcc g++ binutils && \
  curl https://codeload.github.com/universal-ctags/ctags/tar.gz/$CTAGS_VERSION | tar xz -C /tmp && \
  cd /tmp/ctags-$CTAGS_VERSION && \
  ./autogen.sh && \
  LDFLAGS=-static ./configure --program-prefix=universal- --enable-json --enable-seccomp && \
  make -j8 && \
  make install && \
  cd && \
  rm -rf /tmp/ctags-$CTAGS_VERSION && \
  apk --no-cache --purge del build-deps

WORKDIR /
COPY .ctags.d /.ctags.d
