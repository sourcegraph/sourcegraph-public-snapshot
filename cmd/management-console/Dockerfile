FROM sourcegraph/builder:13f3b144f4@sha256:d3cab7dd1d78fb2bc6d109ba364456194a26959122733d8b2a126f10ff384a80 as builder

ARG VERSION="unknown"
ENV VERSION ${VERSION}

ARG MANAGEMENT_CONSOLE_PKG="github.com/sourcegraph/sourcegraph/cmd/management-console"
ENV MANAGEMENT_CONSOLE_PKG ${MANAGEMENT_CONSOLE_PKG}

ARG ENTERPRISE="false"
ENV ENTERPRISE ${ENTERPRISE}

ENV OUTPUT_DIR=/output
RUN mkdir -p ${OUTPUT_DIR}

WORKDIR /src
COPY . .

# Set python2.7 as default (for node-sass)
RUN ln -s /usr/bin/python2.7 /usr/bin/python

RUN ./cmd/management-console/pre-build.sh
RUN ./cmd/management-console/build.sh

FROM sourcegraph/alpine:3.9@sha256:e9264d4748e16de961a2b973cc12259dee1d33473633beccb1dfb8a0e62c6459

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL com.sourcegraph.github.url=https://github.com/sourcegraph/sourcegraph/commit/${COMMIT_SHA}

ENV PGDATABASE=sg PGHOST=pgsql PGPORT=5432 PGSSLMODE=disable PGUSER=sg
USER sourcegraph
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/management-console"]
COPY --from=builder /output/usr/local/bin/management-console /usr/local/bin/
