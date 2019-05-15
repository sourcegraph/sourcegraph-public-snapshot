FROM alpine:3.9@sha256:769fddc7cc2f0a1c35abb2f91432e8beecf83916c421420e6a6da9f8975464b6 AS ctags

ENV CTAGS_VERSION=5b4865cc2d4831db9d638a647ff2f5a0dced2133

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
