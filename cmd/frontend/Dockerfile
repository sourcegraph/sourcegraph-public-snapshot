FROM sourcegraph/builder:13f3b144f4@sha256:d3cab7dd1d78fb2bc6d109ba364456194a26959122733d8b2a126f10ff384a80 as builder

ARG VERSION="unknown"
ENV VERSION ${VERSION}

ARG FRONTEND_PKG="github.com/sourcegraph/sourcegraph/cmd/frontend"
ENV FRONTEND_PKG ${FRONTEND_PKG}

ARG ENTERPRISE="false"
ENV ENTERPRISE ${ENTERPRISE}

ARG PRE_BUILD_SCRIPT="cmd/frontend/pre-build.sh"

ENV OUTPUT_DIR=/output
RUN mkdir -p ${OUTPUT_DIR}

# Set the default python version to 2.7
# (for sourcegraph-frontend's node-sass depdendency)
RUN ln -s /usr/bin/python2.7 /usr/bin/python

WORKDIR /src
COPY . .

RUN ${PRE_BUILD_SCRIPT}
RUN cmd/frontend/build.sh

FROM sourcegraph/alpine:3.9@sha256:e9264d4748e16de961a2b973cc12259dee1d33473633beccb1dfb8a0e62c6459

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

ENV CONFIGURATION_MODE=server PGDATABASE=sg PGHOST=pgsql PGPORT=5432 PGSSLMODE=disable PGUSER=sg PUBLIC_REPO_REDIRECTS=true CACHE_DIR=/mnt/cache/frontend
USER sourcegraph
CMD ["serve"]
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/frontend"]
COPY --from=builder /output/frontend /usr/local/bin/
