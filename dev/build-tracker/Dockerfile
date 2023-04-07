FROM golang:1.19.8-alpine@sha256:841c160ed35923d96c95c52403c4e6db5decd9cbce034aa851e412ade5d4b74f AS build-tracker-build

ENV GO111MODULE on
ENV GOARCH amd64
ENV GOOS linux

COPY . /repo

WORKDIR /repo/dev/build-tracker

RUN go build -o /build-tracker .

FROM sourcegraph/alpine-3.14:211375_2023-04-02_5.0-05ccd1762636@sha256:cd5089e0b0f7e5a5cd18996f5acd23cfa2bde33f6b7ffb9ace47f552941f0be2 AS build-tracker

RUN apk --no-cache add tzdata
COPY --from=build-tracker-build /build-tracker /usr/local/bin/build-tracker
ENTRYPOINT ["build-tracker"]
