# Docker image for Ctags Language Server

FROM alpine:3.4

RUN apk add --update alpine-sdk autoconf automake ca-certificates

RUN mkdir /pt
WORKDIR /pt
RUN wget https://github.com/monochromegane/the_platinum_searcher/releases/download/v2.1.4/pt_linux_386.tar.gz
RUN tar -xvzf pt_linux_386.tar.gz
RUN chmod +x ./pt_linux_386/pt
RUN mv ./pt_linux_386/pt /usr/local/bin/pt

RUN git clone https://github.com/universal-ctags/ctags /ctags
WORKDIR /ctags
RUN ./autogen.sh
RUN ./configure
RUN make
RUN make install

ADD langserver-ctags /langserver-ctags
ENV PATH /:$PATH

ENTRYPOINT ["/langserver-ctags"]
