FROM sourcegraph/alpine-3.14:160198_2022-07-13_453c015a507e@sha256:f056fdd67075680a0299a9442e5778f05cbaa838ae0add9db7387071eb367b76$


ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

USER sourcegraph
ENTRYPOINT ["/usr/local/bin/sg"]
COPY sg /usr/local/bin/
