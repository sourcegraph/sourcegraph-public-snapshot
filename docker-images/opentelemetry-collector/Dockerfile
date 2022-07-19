ARG OTEL_COLLECTOR_VERSION

FROM golang:1.18.1-alpine@sha256:42d35674864fbb577594b60b84ddfba1be52b4d4298c961b46ba95e9fb4712e8 AS healthcheck-build
COPY ./cmd /cmd
WORKDIR /cmd
RUN go mod init github.com/sourcegraph/sourcegraph/docker-images/opentelemetry/cmd/healthcheck
RUN go build -o /bin/healthcheck ./healthcheck

FROM otel/opentelemetry-collector:${OTEL_COLLECTOR_VERSION} as otel-collector

FROM sourcegraph/alpine-3.14:159028_2022-07-07_1f3b17ce1db8@sha256:25d682b5fd069c716c2b29dcf757c0dc0ce29810a07f91e1347901920272b4a7
COPY --from=healthcheck-build /bin/healthcheck /bin
# this doesnt work lol
