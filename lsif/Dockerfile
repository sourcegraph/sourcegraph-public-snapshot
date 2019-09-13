FROM alpine:3.9@sha256:644fcb1a676b5165371437feaa922943aaf7afcfa8bfee4472f6860aad1ef2a0 AS builder

RUN apk add --no-cache nodejs-current=11.3.0-r0 nodejs-npm=10.14.2-r0
RUN npm install -g yarn@1.17.3

COPY package.json yarn.lock /lsif/
RUN yarn --cwd /lsif

COPY tsconfig.json /lsif
COPY src /lsif/src
RUN yarn --cwd /lsif run build

FROM alpine:3.9@sha256:644fcb1a676b5165371437feaa922943aaf7afcfa8bfee4472f6860aad1ef2a0

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

# hadolint ignore=DL3018
RUN apk add --no-cache tini nodejs-current=11.3.0-r0

COPY --from=builder /lsif /lsif

EXPOSE 3186
ENTRYPOINT ["/sbin/tini", "--", "node", "/lsif/out/server.js"]
