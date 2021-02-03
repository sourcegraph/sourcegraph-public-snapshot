FROM sourcegraph/alpine:3.12@sha256:133a0a767b836cf86a011101995641cf1b5cbefb3dd212d78d7be145adde636d

# needed for `src lsif upload` and `src actions exec`
RUN apk add --no-cache git

COPY src /usr/bin/
ENTRYPOINT ["/usr/bin/src"]
