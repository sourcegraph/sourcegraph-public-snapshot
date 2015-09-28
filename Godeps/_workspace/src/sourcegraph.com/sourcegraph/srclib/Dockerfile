FROM ubuntu:14.04

RUN apt-get update -qq
RUN apt-get install -qq golang build-essential git mercurial

ENV GOPATH /srclib
ENV PATH /srclib/bin:$PATH

ADD . /srclib/src/sourcegraph.com/sourcegraph/srclib/
RUN go get -v sourcegraph.com/sourcegraph/srclib/...
RUN go install sourcegraph.com/sourcegraph/srclib/cmd/src
RUN go get -v sourcegraph.com/sourcegraph/srclib-go
RUN cd /srclib/src/sourcegraph.com/sourcegraph && git clone https://sourcegraph.com/sourcegraph/srclib-javascript

RUN mkdir -p /root/.srclib/sourcegraph.com/sourcegraph
RUN ln -rs /srclib/src/sourcegraph.com/sourcegraph/srclib-go /root/.srclib/sourcegraph.com/sourcegraph/srclib-go
RUN ln -rs /srclib/src/sourcegraph.com/sourcegraph/srclib-javascript /root/.srclib/sourcegraph.com/sourcegraph/srclib-javascript

ENTRYPOINT ["/srclib/bin/src"]
