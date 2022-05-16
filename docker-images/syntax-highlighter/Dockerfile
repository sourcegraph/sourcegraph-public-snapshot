###################################
# Build syntect_server statically #
###################################

FROM rust:1.58-alpine3.14@sha256:fbc0cfaa6261179ca7653cc2b20ca9c3e936f3416a911a27a28694f8fb23b8b4 as ss
RUN apk add --no-cache musl-dev>=1.1.24-r10
COPY . /repo
WORKDIR /repo
RUN cargo test --release --workspace
RUN cargo rustc --release
RUN ls ./target
RUN cp ./target/release/syntect_server /syntect_server

################################
# Build http-server-stabilizer #
################################
FROM golang:1.15.2-alpine@sha256:fc801399d044a8e01f125eeb5aa3f160a0d12d6e03ba17a1d0b22ce50dfede81 as hss

RUN apk add --no-cache git>=2.26.3
RUN git clone https://github.com/slimsag/http-server-stabilizer /repo
WORKDIR /repo
RUN git checkout v1.0.4 && go build -o /http-server-stabilizer .

#######################
# Compile final image #
#######################
FROM sourcegraph/alpine-3.14:142406_2022-04-14_8836ac3499f4@sha256:2a2d1cbaec78882661fe1aa5b0a4af0c23a37be2ea9ff8aadc2da5b80852c233
COPY --from=ss syntect_server /
COPY --from=hss http-server-stabilizer /

EXPOSE 9238
ENV ROCKET_ENV "production"
ENV ROCKET_LIMITS "{json=10485760}"

# syntect_server does not need a secret key since it uses no cookies, but
# without one set Rocket emits a warning.
ENV ROCKET_SECRET_KEY "SeerutKeyIsI7releuantAndknvsuZPluaseIgnorYA="

# When keep-alive is on, we observe connection resets in our Go clients of
# syntect_server. It is unclear why this is, especially because our Go clients do
# not reuse the connection (i.e. we make a fresh connection every time).
# Disabling keep-alive does resolve the issue though, our best guess is that
# this is a bug in Hyper 0.10 (see https://github.com/SergioBenitez/Rocket/issues/928#issuecomment-464632953).
# See https://github.com/sourcegraph/sourcegraph/issues/2615 for details on
# what we observed when this was enabled with the default 5s.
ENV ROCKET_KEEP_ALIVE=0

# The more workers, the more resilient syntect_server is to getting stuck on
# bad grammar/file combinations. If it happens with four workers, only 1/4th of
# requests will be affected for a short period of time. Each worker can require
# at peak around 1.1 GiB of memory.
ENV WORKERS=4

ENV QUIET=true
CMD ["sh", "-c", "/http-server-stabilizer -listen=:9238 -prometheus-app-name=syntect_server -workers=$WORKERS -- env ROCKET_PORT={{.Port}} /syntect_server"]
