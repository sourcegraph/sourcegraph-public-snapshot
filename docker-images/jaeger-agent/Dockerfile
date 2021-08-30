# This Dockerfile tracks jaegertracing/jaeger-agent, but includes more debugging tools and runs as a
# non-root user. It requires JAEGER_VERSION to be set as an argument to build.
ARG JAEGER_VERSION
FROM jaegertracing/jaeger-agent:${JAEGER_VERSION} as base

FROM sourcegraph/alpine-3.12:106643_2021-08-30_7f4b12a@sha256:5176711adb03e08bed6f231255aa408f1e1dc67429346e6d5a6811e352e06b88
USER root
# hadolint ignore=DL3018
RUN apk --no-cache add bash curl apk-tools=2.10.8-r0

COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=base /go/bin/agent-linux /go/bin/agent-linux

RUN adduser -S -u 10001 jaeger
RUN chown jaeger /go/bin/agent-linux
USER jaeger

EXPOSE 5775/udp 6831/udp 6832/udp 5778
ENTRYPOINT ["/go/bin/agent-linux"]

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://docs.sourcegraph.com/
LABEL com.sourcegraph.jaeger.version=${JAEGER_VERSION}
