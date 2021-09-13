###################################
# Build syntect_server statically #
###################################
FROM rust:1.46.0-alpine3.12@sha256:c9890db1527309a88994b12dfe5e1ed695518037367ffb74df76a66bfc303ef5 as ss
RUN apk add --no-cache musl-dev=1.1.24-r10
COPY . /repo
WORKDIR /repo
RUN cargo rustc --release
RUN ls ./target
RUN cp ./target/release/syntect_server /syntect_server

################################
# Build http-server-stabilizer #
################################
FROM golang:1.15.2-alpine@sha256:4d8abd16b03209b30b48f69a2e10347aacf7ce65d8f9f685e8c3e20a512234d9 as hss

RUN apk add --no-cache git=2.26.3-r0
RUN git clone https://github.com/slimsag/http-server-stabilizer /repo
WORKDIR /repo
RUN git checkout v1.0.4 && go build -o /http-server-stabilizer .

#######################
# Compile final image #
#######################
FROM sourcegraph/alpine-3.12:106643_2021-08-30_7f4b12a@sha256:5176711adb03e08bed6f231255aa408f1e1dc67429346e6d5a6811e352e06b88
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
