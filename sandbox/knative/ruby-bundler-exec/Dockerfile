FROM golang:1.13 as builder
WORKDIR /work
COPY go.* ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o server

FROM sqs1/a8n-base:latest
USER root
RUN apt-get update && \
    apt-get install -y libssl-dev
RUN cd /tmp && \
    curl -L https://github.com/postmodern/ruby-install/archive/v0.7.0.tar.gz > ruby-install-0.7.0.tar.gz && \
    tar -xzvf ruby-install-0.7.0.tar.gz && \
    cd ruby-install-0.7.0/ && \
    make install && \
    cd /tmp && \
    rm -rf ruby*
ARG RUBY_VERSION
RUN ruby-install -c -j4 --system "ruby-${RUBY_VERSION}" -- --disable-install-doc
RUN chmod -R a+rw /usr/local
ADD gemrc /home/ubuntu/.gemrc
RUN gem install bundler --version '~> 2'
RUN chown -R ubuntu:ubuntu /home/ubuntu
COPY --from=builder /work/server /server
USER ubuntu

CMD ["/server"]
