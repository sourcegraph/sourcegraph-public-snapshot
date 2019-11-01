FROM alpine:3.9@sha256:bf1684a6e3676389ec861c602e97f27b03f14178e5bc3f70dce198f9f160cce9

# hadolint ignore=DL3003,DL3018,DL4006
RUN apk --no-cache add --virtual build-deps curl git gcc make libc-dev pcre-dev sqlite-dev && \
  mkdir /sqlite3-pcre && \
  curl -fsSL https://codeload.github.com/ralight/sqlite3-pcre/tar.gz/c98da412b431edb4db22d3245c99e6c198d49f7a | tar -C /sqlite3-pcre -xzvf - --strip 1 && \
  cd /sqlite3-pcre && \
  make && \
  apk --no-cache --purge del build-deps
