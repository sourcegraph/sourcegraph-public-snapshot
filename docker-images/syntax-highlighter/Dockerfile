# syntax=docker/dockerfile:1.3-labs

###################################
# Build syntect_server statically #
###################################

# Keep Rust version in sync with /WORKSPACE
# and docker-images/syntax-highlighter/rust-toolchain.toml
FROM rust:1.68.0-alpine3.17@sha256:d119a621ae12f84ec0c5fed77c24795120ed1c7874b2428b5a6ccc0f294dbe18 as ss
RUN apk add --no-cache musl-dev>=1.1.24-r10 build-base
COPY . /repo
WORKDIR /repo
RUN cargo fetch

ARG TARGETARCH

# Because .cargo/config.toml doesnt support triplet-specific env
RUN <<EOF
if [ "${TARGETARCH}" = "arm64" ]; then
	cat <<- FOE >> .cargo/config.toml
	[env]
	CFLAGS="-mno-outline-atomics"
	FOE
fi;
EOF

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
FROM sourcegraph/alpine-3.14:213466_2023-04-17_5.0-bdda34a71619@sha256:6354a4ff578b685e36c8fbde81f62125ae0011b047fb2cc22d1b0de616b3c59a
COPY --from=ss syntax_highlighter /
COPY --from=hss http-server-stabilizer /

# even with everything, we can't get this to statically link
RUN apk add --no-cache libstdc++

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
