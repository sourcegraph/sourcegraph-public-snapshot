FROM golang:1.18.1-alpine@sha256:42d35674864fbb577594b60b84ddfba1be52b4d4298c961b46ba95e9fb4712e8 AS build-tracker-build

ENV GO111MODULE on
ENV GOARCH amd64
ENV GOOS linux

COPY . /repo

WORKDIR /repo/dev/build-tracker

RUN go build -o /build-tracker .

FROM sourcegraph/alpine-3.14:154143_2022-06-13_1eababf8817e@sha256:f1c4ac9ca1a36257c1eb699d0acf489d83dd86e067b1fc3ea4a563231a047e05 AS build-tracker

RUN apk --no-cache add tzdata
COPY --from=build-tracker-build /build-tracker /usr/local/bin/build-tracker
ENTRYPOINT ["build-tracker"]
