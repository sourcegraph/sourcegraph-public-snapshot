FROM alpine:3.8
RUN apk add --no-cache bind-tools ca-certificates mailcap tini
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/creaper"]
ADD creaper /usr/local/bin/
