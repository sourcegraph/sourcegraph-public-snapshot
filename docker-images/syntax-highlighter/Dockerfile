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

RUN cp ./target/release/syntect_server /syntax_highlighter

################################
# Build http-server-stabilizer #
################################
FROM golang:1.19-alpine@sha256:70df3b8f9f099da7f60f0b32480015165e3d0b51bfacf9e255b59f3dd6bd2828 as hss

RUN apk add --no-cache git>=2.26.3
RUN git clone --branch v1.1.0 --single-branch https://github.com/sourcegraph/http-server-stabilizer /repo
WORKDIR /repo
RUN go build -o /http-server-stabilizer .

#######################
# Compile final image #
#######################
FROM sourcegraph/alpine-3.14:196830_2023-02-01_af83eee939ca@sha256:b4d7040d41fcf37fbf96fe5a14c39ae15580a3a6c76355cc7ea04a74b6c3b9fa
COPY --from=ss syntax_highlighter /
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
CMD ["sh", "-c", "/http-server-stabilizer -listen=:9238 -prometheus-app-name=syntax_highlighter -workers=$WORKERS -- env ROCKET_PORT={{.Port}} /syntax_highlighter"]
