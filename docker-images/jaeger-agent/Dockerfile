# This Dockerfile tracks jaegertracing/jaeger-agent, but includes more debugging tools and runs as a
# non-root user. It requires JAEGER_VERSION to be set as an argument to build.
ARG JAEGER_VERSION
FROM jaegertracing/jaeger-agent:${JAEGER_VERSION} as base

FROM sourcegraph/alpine-3.14:142406_2022-04-14_8836ac3499f4@sha256:2a2d1cbaec78882661fe1aa5b0a4af0c23a37be2ea9ff8aadc2da5b80852c233
USER root
RUN apk --no-cache add bash curl apk-tools>=2.10.8-r0

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
