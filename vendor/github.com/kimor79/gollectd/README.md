gollectd
========

This is yet another implementation of a [collectd binary protocol](https://collectd.org/wiki/index.php/Binary_protocol) parser in [Go](http://golang.org/), heavenly inspired by [gocollectd](https://github.com/paulhammond/gocollectd).

Installation
------------

`go get github.com/kimor79/gollectd`

Usage
-----

```
import (
    collectd github.com/kimor79/gollectd
)

types, err := collectd.TypesDBFile("/path/to/types.db")

buffer := make([]byte, 1452)
n, _, err := socket.ReadFromUDP(buffer)
packets, err := collectd.Packets(buffer[:n], types)
```
