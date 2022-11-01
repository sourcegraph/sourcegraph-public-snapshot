FROM golang:1.18.1-alpine@sha256:42d35674864fbb577594b60b84ddfba1be52b4d4298c961b46ba95e9fb4712e8 AS build-tracker-build

ENV GO111MODULE on
ENV GOARCH amd64
ENV GOOS linux

COPY . /repo

WORKDIR /repo/dev/build-tracker

RUN go build -o /build-tracker .

FROM sourcegraph/alpine-3.14:180512_2022-10-31_84d1e240bb40@sha256:179ad53ab463ebc804f93de967113739fa73efc2cea6d9c53a9106be45f79d5e AS build-tracker

RUN apk --no-cache add tzdata
COPY --from=build-tracker-build /build-tracker /usr/local/bin/build-tracker
ENTRYPOINT ["build-tracker"]
