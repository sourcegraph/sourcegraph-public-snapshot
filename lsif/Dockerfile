FROM alpine:3.10@sha256:72c42ed48c3a2db31b7dafe17d275b634664a708d901ec9fd57b1529280f01fb AS builder

RUN apk add --no-cache nodejs-current=12.4.0-r0 nodejs-npm=10.16.3-r0
RUN npm install -g yarn@1.17.3

COPY . /lsif/
RUN yarn --cwd /lsif
RUN yarn --cwd /lsif run build

FROM alpine:3.9@sha256:72c42ed48c3a2db31b7dafe17d275b634664a708d901ec9fd57b1529280f01fb

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

# hadolint ignore=DL3018
RUN apk add --no-cache tini nodejs-current=12.4.0-r0

COPY --from=builder /lsif /lsif

EXPOSE 3186
ENTRYPOINT ["/sbin/tini", "--", "node", "/lsif/out/server.js"]
