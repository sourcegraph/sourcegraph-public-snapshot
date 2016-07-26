FROM alpine:3.4

ADD guru /guru
ADD langserver-go /langserver-go

# HACK: clone the Sourcegraph repository for testing purposes until we figure
# out correct FS architecture.
RUN apk add --update git
RUN git clone https://github.com/sourcegraph/sourcegraph /gopath/src/github.com/sourcegraph/sourcegraph

ENTRYPOINT ["/langserver-go"]
