# Docker image for the srclib importer.
#
#     docker build -t sourcegraph/srclib-import - < srclib-import.dockerfile
#     docker push sourcegraph/srclib-import
#
# URL: https://hub.docker.com/r/sourcegraph/srclib-import/

FROM alpine:3.2
MAINTAINER Sourcegraph Team <help@sourcegraph.com>

RUN apk add -q -U ca-certificates zip curl && rm -rf /var/cache/apk/*
