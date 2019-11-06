# libsqlite3-pcre must be built using Alpine (has musl) instead of Ubuntu (has glibc) (which sourcegraph/builder is based on.).
# Since sourcegraph/server uses Alpine, libsqlite3-pcre must be built using musl. I was not able to figure out how to use Ubuntu
# to cross-compile libsqlite3-pcre in a way that worked for Alpine (forcing "CC=musl-gcc" wasn't sufficient).

FROM alpine:3.9@sha256:644fcb1a676b5165371437feaa922943aaf7afcfa8bfee4472f6860aad1ef2a0 as builder

ENV OUTPUT_DIR=/output
RUN mkdir -p ${OUTPUT_DIR}

# hadolint ignore=DL3018
RUN apk --no-cache add --virtual build-deps \
    bash \
    curl \
    git \
    gcc \
    make \
    libc-dev \
    pcre-dev \
    sqlite-dev

WORKDIR /src
COPY . .

# Alpine-linux already uses musl by default
ENV CC "gcc"

RUN ./build.sh

FROM alpine:3.10@sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a

# hadolint ignore=DL3018
RUN apk --no-cache add pcre-dev

COPY --from=builder /output/pcre.so /libsqlite3-pcre.so
ENV LIBSQLITE3_PCRE /libsqlite3-pcre.so
