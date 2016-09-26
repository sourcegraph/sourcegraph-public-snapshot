# Docker image for LSP proxy

FROM alpine:3.4
RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/community ca-certificates
ADD lsp-proxy /lsp-proxy
ENTRYPOINT ["/lsp-proxy"]
