# This Dockerfile was generated from github.com/sourcegraph/godockerize. It
# was not written by a human, and as such looks janky. As you change this
# file, please don't be scared to make it more pleasant / remove hadolint
# ignores.

FROM sourcegraph/alpine:3.9
ENV PGDATABASE=sg PGHOST=pgsql PGPORT=5432 PGSSLMODE=disable PGUSER=sg
USER sourcegraph
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/management-console"]
COPY management-console /usr/local/bin/
