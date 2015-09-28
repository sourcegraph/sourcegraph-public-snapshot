FROM ubuntu:14.04

RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    rpm \
    ruby-dev
RUN gem install fpm

WORKDIR /release

ADD scripts scripts
RUN /release/scripts/install_daemonize.sh bin

ADD . /release

ENTRYPOINT ["make"]
