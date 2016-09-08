# Docker image for Ctags Language Server

FROM alpine:3.4

RUN apk add --update alpine-sdk
RUN apk add autoconf
RUN apk add automake

RUN git clone https://github.com/universal-ctags/ctags /ctags
WORKDIR /ctags
RUN ./autogen.sh
RUN ./configure
RUN make
RUN make install

ADD langserver-ctags /langserver-ctags
ENV PATH /:$PATH

ENTRYPOINT ["/langserver-ctags"]
