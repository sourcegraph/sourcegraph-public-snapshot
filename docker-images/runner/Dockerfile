# TODO: This file should probably go into the sourcegraph/src-cli repo

FROM sourcegraph/alpine:latest

RUN apk update && apk add --no-cache docker git unzip diffutils

USER sourcegraph

CMD ["actions", "runner"]

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/src-cli"]

COPY --chown=sourcegraph:sourcegraph ./src /usr/local/bin/src-cli
