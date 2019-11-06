FROM sourcegraph/builder:1be4c9ba43@sha256:0172c10814a1fc71adfe82c19c69bcf3f1496745c99e865e32836466b148778b as builder

ENV OUTPUT_DIR=/output
RUN mkdir -p ${OUTPUT_DIR}

WORKDIR /src
COPY . .

ARG CTAGS_VERSION="03f933a96d3ef87adbf9d167462d45ce69577edb"
ENV CTAGS_VERSION ${CTAGS_VERSION}

ARG BUILD_TYPE="dev"
ENV BUILD_TYPE ${BUILD_TYPE}

ARG VERSION="unknown"
ENV VERSION ${VERSION}

ENV CTAGS_D_OUTPUT_PATH ${OUTPUT_DIR}/.ctags.d
ENV SYMBOLS_EXECUTABLE_OUTPUT_PATH=${OUTPUT_DIR}/symbols

RUN cmd/symbols/build.sh build_binary_and_universal_ctags

# libsqlite3-pcre must be built using Alpine (has musl) instead of Ubuntu (has glibc) (which sourcegraph/builder is based on.).
# Since sourcegraph/symbols uses Alpine, libsqlite3-pcre must be built using musl. I was not able to figure out how to use Ubuntu
# to cross-compile libsqlite3-pcre in a way that worked for Alpine (forcing "CC=musl-gcc" wasn't sufficient).
FROM alpine:3.10@sha256:72c42ed48c3a2db31b7dafe17d275b634664a708d901ec9fd57b1529280f01fb AS libsqlite3-pcre-builder

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

# Alpine-linux already uses musl by default
ENV CC "gcc"

WORKDIR /src
COPY cmd/symbols/libsqlite3-pcre .
RUN ./build.sh

FROM alpine:3.10@sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

# hadolint ignore=DL3018
RUN apk add --no-cache bind-tools ca-certificates mailcap tini

# hadolint ignore=DL3022
COPY --from=builder /output/bin/universal-* /usr/local/bin/

# hadolint ignore=DL3022
COPY --from=libsqlite3-pcre-builder /output/pcre.so /libsqlite3-pcre.so
ENV LIBSQLITE3_PCRE /libsqlite3-pcre.so
# hadolint ignore=DL3018
RUN apk --no-cache add pcre-dev

# universal-ctags doesn't support some popular languages out of the box (Scala,
# Swift, and Haskell, to name a few). However, you can add language support by
# providing regexes on the command line or in a configuration file.
# universal-ctags looks in the current working directory for config files during
# initialization:
# https://github.com/universal-ctags/ctags/blob/1b747b38589f7a1c98ac9fe0a0b8b174dd476050/docs/optlib.rst#preload-option-file
WORKDIR /
COPY cmd/symbols/.ctags.d /.ctags.d

ENV CACHE_DIR=/mnt/cache/symbols
EXPOSE 3184
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/symbols"]
COPY --from=builder /output/symbols /usr/local/bin/
