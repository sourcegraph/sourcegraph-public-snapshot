# Docker image for the srclib importer.
#
#     docker build -t sourcegraph/srclib-import .
#     docker push sourcegraph/srclib-import
#
# URL: https://hub.docker.com/r/sourcegraph/srclib-import/

FROM alpine:3.3
MAINTAINER Sourcegraph Team <support@sourcegraph.com>

RUN apk add -q -U ca-certificates zip && rm -rf /var/cache/apk/*

ADD srclib-import /usr/bin/
