# godockerize

godockerize builds a Dockerfile for a given Go project. The base image uses
Alpine.

### Usage

```
$ godockerize build github.com/path/to/your/go/main
```

Will generate a Dockerfile, build a Go binary for that package, and then build
the Dockerfile with the Go binary.

Run `--help` to view the full list of options for customizing the program's
behavior.

### Build tags

godockerize supports a number of build tags that can be used to customize the
content of the generated Dockerfile. Add a `//docker:<tagname>` comment line to
the package being built to customize the behavior. Here are the tags:

##### //docker:env

Adding this to your Go binary:

```go
//docker:env TZ=UTC

package main
```

Adds an ENV command to the dockerfile:

```
ENV TZ=UTC
```

##### //docker:expose

Expose the provided ports via an EXPOSE command

```go
//docker:expose 5000
```

##### //docker:install

Install the provided packages. In addition, if a package name ends with `@edge`,
the edge apk repositories will be added to `/etc/apk/repositories`.

```go
//docker:install git@edge openssh-client
```

Generates:

```
  FROM alpine:3.8
  RUN echo -e "@edge http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories
  RUN echo -e "@edge http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories
  RUN apk add --no-cache ca-certificates git@edge mailcap openssh-client tini
  USER myuser
  ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/mybinary"]
  ADD mybinary /usr/local/bin/
```

##### //docker:repository

Can be used to add other repository versions. Note as a special case, the `edge`
repository is automatically added as part of the `docker:install` step for any
packages that end in `@edge`.

```go
//docker:repository v3.6
```

Will add these lines to the Dockerfile:

```
  RUN echo -e "http://dl-cdn.alpinelinux.org/alpine/v3.6/main\n" >> /etc/apk/repositories && \
    echo -e "http://dl-cdn.alpinelinux.org/alpine/v3.6/community\n" >> /etc/apk/repositories
```

##### //docker:run

Add custom commands to run in the Dockerfile.

```go
//docker:run apk add curl
```

##### //docker:user

Run commands in the Dockerfile as a custom user.

```go
//docker:user myuser
```

```
RUN addgroup -S myuser && adduser -S -G myuser -h /home/myuser myuser && mkdir -p /data/repos && chown -R myuser:myuser /data/repos
USER myuser
```

### Errata

Alpine base images had a security vulnerability where apk installing/unpacking
resources onto the filesystem could lead to RCE. That vulnerability has been
fixed in Alpine 3.8.1. `godockerize` uses a base image that includes the
security patch.
