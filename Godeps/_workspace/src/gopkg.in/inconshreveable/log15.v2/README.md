![obligatory xkcd](http://imgs.xkcd.com/comics/standards.png)

# log15 [![godoc reference](https://godoc.org/gopkg.in/inconshreveable/log15.v2?status.png)](https://godoc.org/gopkg.in/inconshreveable/log15.v2)

Package log15 provides an opinionated, simple toolkit for best-practice logging in Go (golang) that is both human and machine readable. It is modeled after the Go standard library's [`io`](http://golang.org/pkg/io/) and [`net/http`](http://golang.org/pkg/net/http/) packages and is an alternative to the standard library's [`log`](http://golang.org/pkg/log/) package. 

## Features
- A simple, easy-to-understand API
- Promotes structured logging by encouraging use of key/value pairs
- Child loggers which inherit and add their own private context
- Lazy evaluation of expensive operations
- Simple Handler interface allowing for construction of flexible, custom logging configurations with a tiny API.
- Color terminal support
- Built-in support for logging to files, streams, syslog, and the network
- Support for forking records to multiple handlers, buffering records for output, failing over from failed handler writes, + more

## Versioning
The API of the master branch of log15 should always be considered unstable. Using a stable version
of the log15 package is supported by gopkg.in. Include your dependency like so:

```go
import log "gopkg.in/inconshreveable/log15.v2"
```

## Examples

```go
// all loggers can have key/value context
srvlog := log.New("module", "app/server")

// all log messages can have key/value context 
srvlog.Warn("abnormal conn rate", "rate", curRate, "low", lowRate, "high", highRate)

// child loggers with inherited context
connlog := srvlog.New("raddr", c.RemoteAddr())
connlog.Info("connection open")

// lazy evaluation
connlog.Debug("ping remote", "latency", log.Lazy(pingRemote))

// flexible configuration
srvlog.SetHandler(log.MultiHandler(
    log.StreamHandler(os.Stderr, log.LogfmtFormat()),
    log.LvlFilterHandler(
        log.LvlError,
        log.Must.FileHandler("errors.json", log.JsonHandler())))
```

## FAQ

### The varargs style is brittle and error prone! Can I have type safety please?
Yes. Use `log.Ctx`:

```go
srvlog := log.New(log.Ctx{"module": "app/server"})
srvlog.Warn("abnormal conn rate", log.Ctx{"rate": curRate, "low": lowRate, "high": highRate})
```

## License
Apache
