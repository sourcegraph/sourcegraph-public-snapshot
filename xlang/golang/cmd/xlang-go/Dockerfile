# Docker image for Go lang/build server

FROM alpine:3.4
RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/community ca-certificates
ADD xlang-go /xlang-go
ENTRYPOINT ["/xlang-go"]
