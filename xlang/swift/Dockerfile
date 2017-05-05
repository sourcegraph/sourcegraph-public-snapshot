# Swift 3.1 not yet supported by SourceKitten
FROM norionomura/sourcekit:302

# === Go dependency, copied from: https://github.com/docker-library/golang/blob/07253735ad448adf751e4f9ba4c71acdb735367f/1.8/Dockerfile 
# gcc for cgo
RUN apt-get update && apt-get install -y --no-install-recommends \
		g++ \
		wget \
		gcc \
		libc6-dev \
		make \
		pkg-config \
	&& rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.8.1

RUN set -eux; \
	\
# this "case" statement is generated via "update.sh"
	dpkgArch="$(dpkg --print-architecture)"; \
	case "${dpkgArch##*-}" in \
		ppc64el) goRelArch='linux-ppc64le'; goRelSha256='b7b47572a2676449716865a66901090c057f6f1d8dfb1e19528fcd0372e5ce74' ;; \
		i386) goRelArch='linux-386'; goRelSha256='cb3f4527112075a8b045d708f793aeee2709d2f5ddd320973a1413db06fddb50' ;; \
		s390x) goRelArch='linux-s390x'; goRelSha256='0a59f4034a27fc51431989da520fd244d5261f364888134cab737e5bc2158cb2' ;; \
		armhf) goRelArch='linux-armv6l'; goRelSha256='e8a8326913640409028ef95c2107773f989b1b2a6e11ceb463c77c42887381da' ;; \
		amd64) goRelArch='linux-amd64'; goRelSha256='a579ab19d5237e263254f1eac5352efcf1d70b9dacadb6d6bb12b0911ede8994' ;; \
		*) goRelArch='src'; goRelSha256='33daf4c03f86120fdfdc66bddf6bfff4661c7ca11c5da473e537f4d69b470e57'; \
			echo >&2; echo >&2 "warning: current architecture ($dpkgArch) does not have a corresponding Go binary release; will be building from source"; echo >&2 ;; \
	esac; \
	\
	url="https://golang.org/dl/go${GOLANG_VERSION}.${goRelArch}.tar.gz"; \
	wget -O go.tgz "$url"; \
	echo "${goRelSha256} *go.tgz" | sha256sum -c -; \
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	\
	if [ "$goRelArch" = 'src' ]; then \
		echo >&2; \
		echo >&2 'error: UNIMPLEMENTED'; \
		echo >&2 'TODO install golang-any from jessie-backports for GOROOT_BOOTSTRAP (and uninstall after build)'; \
		echo >&2; \
		exit 1; \
	fi; \
	\
	export PATH="/usr/local/go/bin:$PATH"; \
	go version

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
# ==== End Go dependency

# Build SourceKitten
WORKDIR /
RUN git clone https://github.com/jpsim/SourceKitten.git /SourceKitten
WORKDIR /SourceKitten
RUN swift build
RUN mv ./.build/debug/* /usr/local/bin

# Build Swift Langserver
ADD swift-langserver $GOPATH/src/github.com/sourcegraph/swift-langserver
WORKDIR $GOPATH/src/github.com/sourcegraph/swift-langserver
RUN pwd
RUN go get -d ./...
RUN go install ./

ENV X_REMOTE_FS_ENABLE=t
EXPOSE 2089
CMD ["swift-langserver", "--mode", "tcp", "--addr", ":2089"]

