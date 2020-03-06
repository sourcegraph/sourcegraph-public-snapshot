FROM alpine:3.10@sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a AS lsif-builder

RUN apk add --no-cache nodejs-current=12.4.0-r0 nodejs-npm=10.19.0-r0
RUN npm install -g yarn@1.17.3

COPY lsif/package.json lsif/tsconfig.json lsif/yarn.lock /lsif/
COPY lsif/src /lsif/src

RUN yarn --cwd /lsif
RUN yarn --cwd /lsif run build

# TODO: Make this image use our sourcegraph/alpine:3.10@sha256:4d05cd5669726fc38823e92320659a6d1ef7879e62268adec5df658a0bacf65c
FROM alpine:3.10@sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

# set default environment variables
ENV PGDATABASE=sg PGHOST=pgsql PGPORT=5432 PGSSLMODE=disable PGUSER=sg

# hadolint ignore=DL3018
RUN apk update && apk add --no-cache \
    tini nodejs-current=12.4.0-r0

COPY --from=lsif-builder /lsif /lsif

COPY . /

EXPOSE 3186 3187
ENV GO111MODULES=on LANG=en_US.utf8
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/lsif-server"]
