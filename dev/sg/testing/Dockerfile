# Build SG
# FROM golang:1.17 as go
# ## Copy the entire code tree
# WORKDIR /go/src/sourcegraph
# COPY . .
# ## Build sg 
# RUN cd dev/sg && \
#   go build -ldflags "-X main.BuildCommit=dev" -o /sg .

# Build the base image
FROM ubuntu:latest
RUN apt-get update && \
  # Must be done first to avoid being triggered by other packages
  DEBIAN_FRONTEND=noninteractive TZ=Etc/UTC apt-get -y install tzdata && \ 
  apt-get install -y sudo lsb-core software-properties-common libev-dev

COPY dev/sg/sg_docker /usr/bin/sg
