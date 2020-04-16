FROM sourcegraph/alpine:3.10

# needed for `src lsif upload` and `src actions exec`
RUN apk add --no-cache git

COPY src /usr/bin/
ENTRYPOINT ["/usr/bin/src"]
