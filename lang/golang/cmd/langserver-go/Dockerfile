FROM alpine:3.4

RUN apk add --no-cache go

ADD guru /guru
ADD langserver-go /langserver-go
ENV GOROOT /usr/lib/go
ENV PATH /:$PATH

ENTRYPOINT ["/langserver-go"]
