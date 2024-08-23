# URLx
[Golang](http://golang.org/) pkg for URL parsing and normalization.

1. [Parsing URL](#parsing-url) ([GoDoc](https://godoc.org/github.com/goware/urlx#Parse))
2. [Normalizing URL](#normalizing-url)  ([GoDoc](https://godoc.org/github.com/goware/urlx#Normalize))
3. [Splitting host:port from URL](#splitting-hostport-from-url) ([GoDoc](https://godoc.org/github.com/goware/urlx#SplitHostPort))
4. [Resolving IP address from URL](#resolving-ip-address-from-url) ([GoDoc](https://godoc.org/github.com/goware/urlx#Resolve))

[![GoDoc](https://godoc.org/github.com/goware/urlx?status.png)](https://godoc.org/github.com/goware/urlx)
[![Travis](https://travis-ci.org/goware/urlx.svg?branch=master)](https://travis-ci.org/goware/urlx)

## Parsing URL

The [urlx.Parse()](https://godoc.org/github.com/goware/urlx#Parse) is compatible with the same function from [net/url](https://golang.org/pkg/net/url/#Parse) pkg, but has slightly different behavior. It enforces default scheme and favors absolute URLs over relative paths.

### Difference between [urlx](https://godoc.org/github.com/goware/urlx#Parse) and [net/url](https://golang.org/pkg/net/url/#Parse)

<table>
<thead>
<tr>
<th><a href="https://godoc.org/github.com/goware/urlx#Parse">github.com/goware/urlx</a></th>
<th><a href="https://golang.org/pkg/net/url/#Parse">net/url</a></th>
</tr>
</thead>
<tr>
<td>
<pre>
urlx.Parse("example.com")

&url.URL{
   Scheme:  "http",
   Host:    "example.com",
   Path:    "",
}
</pre>
</td>
<td>
<pre>
url.Parse("example.com")

&url.URL{
   Scheme:  "",
   Host:    "",
   Path:    "example.com",
}
</pre>
</td>
</tr>
<tr>
<td>
<pre>
urlx.Parse("localhost:8080")

&url.URL{
   Scheme:  "http",
   Host:    "localhost:8080",
   Path:    "",
   Opaque:  "",
}
</pre>
</td>
<td>
<pre>
url.Parse("localhost:8080")

&url.URL{
   Scheme:  "localhost",
   Host:    "",
   Path:    "",
   Opaque:  "8080",
}
</pre>
</td>
</tr>
<tr>
<td>
<pre>
urlx.Parse("user.local:8000/path")

&url.URL{
   Scheme:  "http",
   Host:    "user.local:8000",
   Path:    "/path",
   Opaque:  "",
}
</pre>
</td>
<td>
<pre>
url.Parse("user.local:8000/path")

&url.URL{
   Scheme:  "user.local",
   Host:    "",
   Path:    "",
   Opaque:  "8000/path",
}
</pre>
</td>
</tr>
</table>

### Usage

```go
import "github.com/goware/urlx"

func main() {
    url, _ := urlx.Parse("example.com")
    // url.Scheme == "http"
    // url.Host == "example.com"

    fmt.Print(url)
    // Prints http://example.com
}
```

## Normalizing URL

The [urlx.Normalize()](https://godoc.org/github.com/goware/urlx#Normalize) function normalizes the URL using the predefined subset of [Purell](https://github.com/PuerkitoBio/purell) flags.

### Usage

```go
import "github.com/goware/urlx"

func main() {
    url, _ := urlx.Parse("localhost:80///x///y/z/../././index.html?b=y&a=x#t=20")
    normalized, _ := urlx.Normalize(url)

    fmt.Print(normalized)
    // Prints http://localhost/x/y/index.html?a=x&b=y#t=20
}
```

## Splitting host:port from URL

The [urlx.SplitHostPort()](https://godoc.org/github.com/goware/urlx#SplitHostPort) is compatible with the same function from [net](https://golang.org/pkg/net/) pkg, but has slightly different behavior. It doesn't remove brackets from `[IPv6]` host.

### Usage

```go
import "github.com/goware/urlx"

func main() {
    url, _ := urlx.Parse("localhost:80")
    host, port, _ := urlx.SplitHostPort(url)

    fmt.Print(host)
    // Prints localhost

    fmt.Print(port)
    // Prints 80
}
```

## Resolving IP address from URL

The [urlx.Resolve()](https://godoc.org/github.com/goware/urlx#Resolve) is compatible with [ResolveIPAddr()](https://golang.org/pkg/net/#ResolveIPAddr) from [net](https://golang.org/pkg/net/).

### Usage

```go
url, _ := urlx.Parse("localhost")
ip, _ := urlx.Resolve(url)

fmt.Print(ip)
// Prints 127.0.0.1
```

## License
URLx is licensed under the [MIT License](./LICENSE).
