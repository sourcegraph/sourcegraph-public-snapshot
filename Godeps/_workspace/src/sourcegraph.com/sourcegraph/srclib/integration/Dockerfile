# This Dockerfile installs all of the necessary deps for the standard
# toolchains.

FROM ubuntu:14.04

# I was having trouble downloading from archive.ubuntu.com, so I added this
# line. You probably don't need it.
RUN awk 'BEGIN{print "deb mirror://mirrors.ubuntu.com/mirrors.txt trusty main restricted universe multiverse\ndeb mirror://mirrors.ubuntu.com/mirrors.txt trusty-updates main restricted universe multiverse\ndeb mirror://mirrors.ubuntu.com/mirrors.txt trusty-backports main restricted universe multiverse\ndeb mirror://mirrors.ubuntu.com/mirrors.txt trusty-security main restricted universe multiverse"}{print}' > /etc/apt/sources.list

RUN apt-get update && echo 2014-08-13
RUN apt-get install -qq curl git mercurial build-essential

# Install Go
RUN curl -Lo /tmp/golang.tgz https://storage.googleapis.com/golang/go1.3.1.linux-amd64.tar.gz
RUN tar -xzf /tmp/golang.tgz -C /usr/local
ENV GOROOT /usr/local/go
ENV GOBIN /usr/local/bin
ENV PATH /usr/local/go/bin:$PATH
ENV GOPATH /tmp/gopath

# Install Python
RUN apt-get install -qq python

# Install Node.js
RUN apt-get install -qqy nodejs npm
RUN ln -s /usr/bin/nodejs /usr/bin/node

# Install Ruby
RUN curl -L https://get.rvm.io | bash -s stable
ENV PATH /usr/local/rvm/bin:$PATH
RUN rvm requirements
ENV RUBY_VERSION ruby-2.1.2
RUN rvm install $RUBY_VERSION
RUN rvm $RUBY_VERSION do gem install bundler --no-ri --no-rdoc
ENV PATH /usr/local/rvm/rubies/$RUBY_VERSION/bin:/usr/local/rvm/gems/$RUBY_VERSION/bin:$PATH

# Add srclib (unprivileged) user
RUN useradd -ms /bin/bash srclib
USER srclib
WORKDIR /home/srclib

USER root
RUN chown -R srclib /usr/local/rvm
ADD ./srclib /usr/local/bin/srclib
USER srclib

RUN srclib toolchain install go python ruby javascript java
