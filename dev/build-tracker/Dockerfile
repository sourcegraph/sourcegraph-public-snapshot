FROM golang:1.19.3-alpine@sha256:27a9653759f44afd08c94418307a26d2db9cf78af12933200bc2ca63c4844316 AS build-tracker-build

ENV GO111MODULE on
ENV GOARCH amd64
ENV GOOS linux

COPY . /repo

WORKDIR /repo/dev/build-tracker

RUN go build -o /build-tracker .

FROM sourcegraph/alpine-3.14:196830_2023-02-01_af83eee939ca@sha256:b4d7040d41fcf37fbf96fe5a14c39ae15580a3a6c76355cc7ea04a74b6c3b9fa AS build-tracker

RUN apk --no-cache add tzdata
COPY --from=build-tracker-build /build-tracker /usr/local/bin/build-tracker
ENTRYPOINT ["build-tracker"]
