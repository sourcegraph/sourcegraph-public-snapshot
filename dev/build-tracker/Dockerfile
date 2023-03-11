FROM golang:1.19.6-alpine@sha256:f2e0acaf7c628cd819b73541d7c1ea8f888d51edb0a58935a3c46a084fa953fa AS build-tracker-build

ENV GO111MODULE on
ENV GOARCH amd64
ENV GOOS linux

COPY . /repo

WORKDIR /repo/dev/build-tracker

RUN go build -o /build-tracker .

FROM sourcegraph/alpine-3.14:201280_2023-02-23_4.5-1071f8b97a60@sha256:c4970b21169db155c1b497740e622adb23007ac11a87ec571d9ecef8aba0adc5 AS build-tracker

RUN apk --no-cache add tzdata
COPY --from=build-tracker-build /build-tracker /usr/local/bin/build-tracker
ENTRYPOINT ["build-tracker"]
