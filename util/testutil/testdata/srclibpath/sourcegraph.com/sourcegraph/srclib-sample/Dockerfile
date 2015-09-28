FROM busybox

RUN echo

# Add this toolchain
ADD . /srclib/srclib-sample/
WORKDIR /srclib/srclib-sample
ENV PATH /srclib/srclib-sample/.bin:$PATH

# Add srclib (unprivileged) user
RUN adduser -D -s /bin/bash srclib
RUN mkdir /src
RUN chown -R srclib /src /srclib
USER srclib

WORKDIR /src

ENTRYPOINT ["srclib-sample"]
