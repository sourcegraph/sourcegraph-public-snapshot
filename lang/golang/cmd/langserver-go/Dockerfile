# Docker image for Go Language Server

FROM alpine:3.4

RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/community go

ADD guru /guru
ADD godef /godef
ADD gog /gog
ADD langserver-go /langserver-go
ENV GOROOT /usr/lib/go
ENV PATH /:$PATH

ENTRYPOINT ["/langserver-go"]
