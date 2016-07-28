FROM alpine:3.4

ADD langprocessor-go /langprocessor-go

# HACK: Matches langserver-go/Dockerfile workspace setup.
ENV GOPATH=/gopath

ENTRYPOINT ["/langprocessor-go"]
