FROM golang:1.10-alpine
WORKDIR /go/src/sourcegraph.com/sourcegraph/sourcegraph/xlang/cloneproxy
RUN apk add --no-cache ca-certificates git
COPY . .
RUN CGO_ENABLED=0 go get -d -v ./...
RUN CGO_ENABLED=0 go build -o cloneproxy *.go

# 👀 Add steps here to build the language server itself 👀
CMD ["echo", "🚨 This statement should be removed once you have added the logic to start up the language server! 🚨 Exiting..."]

# Modify these commands to connect to the language server
#COPY --from=0 /go/src/sourcegraph.com/sourcegraph/sourcegraph/xlang/cloneproxy/cloneproxy .
#EXPOSE 8080
#CMD ["./cloneproxy", "--proxyAddress=0.0.0.0:8080", ...]
