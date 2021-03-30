FROM golang AS BUILDER
WORKDIR /build
COPY go.sum go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o searchblitz .

FROM sourcegraph/alpine:3.12@sha256:133a0a767b836cf86a011101995641cf1b5cbefb3dd212d78d7be145adde636d
RUN mkdir data

COPY --from=builder /build/searchblitz /usr/local/bin
COPY data data

ARG COMMIT_SHA="unknown"

LABEL org.opencontainers.image.revision=${COMMIT_SHA}
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/search-blitz/

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/searchblitz"]
