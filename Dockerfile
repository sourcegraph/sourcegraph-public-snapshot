FROM sourcegraph/alpine:3.10@sha256:4d05cd5669726fc38823e92320659a6d1ef7879e62268adec5df658a0bacf65c

# needed for `src lsif upload` and `src actions exec`
RUN apk add --no-cache git

COPY src /usr/bin/
ENTRYPOINT ["/usr/bin/src"]
