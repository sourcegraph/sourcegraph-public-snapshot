# streamline [![go reference](https://pkg.go.dev/badge/go.bobheadxi.dev/streamline.svg)](https://pkg.go.dev/go.bobheadxi.dev/streamline) [![Sourcegraph](https://img.shields.io/badge/view%20on-sourcegraph-A112FE?logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADIAAAAyCAYAAAAeP4ixAAAEZklEQVRoQ+2aXWgUZxSG3292sxtNN43BhBakFPyhxSujRSxiU1pr7SaGXqgUxOIEW0IFkeYighYUxAuLUlq0lrq2iCDpjWtmFVtoG6QVNOCFVShVLyxIk0DVjZLMxt3xTGTccd2ZOd/8JBHci0CY9zvnPPN+/7sCIXwKavOwAcy2QgngQiIztDSE0OwQlDPYR1ebiaH6J5kZChyfW12gRG4QVgGTBfMchMbFP9Sn5nlZL2D0JjLD6710lc+z0NfqSGTXQRQ4bX07Mq423yoBL3OSyHSvUxirMuaEvgbJWrdcvkHMoJwxYuq4INUhyuWvQa1jvdMGxAvCxJlyEC9XOBCWL04wwRzpbDoDQ7wfZJzIQLi5Eggk6DiRhZgWIAbE3NrM4A3LPT8Q7UgqAqLqTmLSHLGPkyzG/qXEczhd0q6RH+zaSBfaUoc4iQx19pIClIscrTkNZzG6gd7qMY6eC2Hqyo705ZfTf+eqJmhMzcSbYtQpOXc92ZsZjLVAL4YNUQbJ5Ttg4CQrQdGYj44Xr9m1XJCzmZusFDJOWNpHjmh5x624a2ZFtOKDVL+uNo2TuXE3bZQQZUf8gtgqP31uI94Z/rMqix+IGiRfWw3xN9dCgVx+L3WrHm4Dju6PXz/EkjuXJ6R+IGgyOE1TbZqTq9y1eo0EZo7oMo1ktPu3xjHvuiLT5AFNszUyDULtWpzE2/fEsey8O5TbWuGWwxrs5rS7nFNMWJrNh2No74s9Ec4vRNmRRzPXMP19fBMSVsGcOJ98G8N3Wl2gXcbTjbX7vUBxLaeASDQCm5Cu/0E2tvtb0Ea+BowtskFD0wvlc6Rf2M+Jx7dTu7ubFr2dnKDRaMQe2v/tcIrNB7FH0O50AcrBaApmRDVwFO31ql3pD8QW4dP0feNwl/Q+kFEtRyIGyaWXnpy1OO0qNJWHo1y6iCmAGkBb/Ru+HenDWIF2mo4r8G+tRRzoniSn2uqFLxANhe9LKHVyTbz6egk9+x5w5fK6ulSNNMhZ/Feno+GebLZV6isTTa6k5qNl5RnZ5u56Ib6SBvFzaWBBVFZzvnERWlt/Cg4l27XChLCqFyLekjhy6xJyoytgjPf7opIB8QPx7sYFiMXHPGt76m741MhCKMZfng0nBOIjmoJPsLqWHwgFpe6V6qtfcopxveR2Oy+J0ntIN/zCWkf8QNAJ7y6d8Bq4lxLc2/qJl5K7t432XwcqX5CrI34gzATWuYILQtdQPyePDK3iuOekCR3Efjhig1B1Uq5UoXEEoZX7d1q535J5S9VOeFyYyEBku5XTMXXKQTToX5Rg7OI44nbW5oKYeYK4EniMeF0YFNSmb+grhc84LyRCEP1/OurOcipCQbKxDeK2V5FcVyIDMQvsgz5gwFhcWWwKyRlvQ3gv29RwWoDYAbIofNyBxI9eDlQ+n3YgsgCWnr4MStGXQXmv9pF2La/k3OccV54JEBM4yp9EsXa/3LfO0dGPcYq0Y7DfZB8nJzZw2rppHgKgVHs8L5wvRwAAAABJRU5ErkJggg==)](https://sourcegraph.com/github.com/bobheadxi/streamline)

[![pipeline](https://github.com/bobheadxi/streamline/actions/workflows/pipeline.yaml/badge.svg)](https://github.com/bobheadxi/streamline/actions/workflows/pipeline.yaml)
[![codecov](https://codecov.io/gh/bobheadxi/streamline/branch/main/graph/badge.svg?token=f1VZULSJsT)](https://codecov.io/gh/bobheadxi/streamline)
[![Go Report Card](https://goreportcard.com/badge/go.bobheadxi.dev/streamline)](https://goreportcard.com/report/go.bobheadxi.dev/streamline)
[![benchmarks](https://img.shields.io/website/https/bobheadxi.dev/streamline.svg?down_color=red&down_message=offline&label=benchmarks&up_message=live)](https://bobheadxi.dev/streamline)

Transform and handle your data, line by line.

```sh
go get go.bobheadxi.dev/streamline
```

## Overview

[`streamline`](https://pkg.go.dev/go.bobheadxi.dev/streamline) offers a variety of primitives to make working with data line by line a breeze:

- [`streamline.Stream`](https://pkg.go.dev/go.bobheadxi.dev/streamline#Stream) offers the ability to add hooks that handle an `io.Reader` line-by-line with `(*Stream).Stream`, `(*Stream).StreamBytes`, and other utilities.
- [`pipeline.Pipeline`](https://pkg.go.dev/go.bobheadxi.dev/streamline/pipeline#Pipeline) offers a way to build pipelines that transform the data in a `streamline.Stream`, such as cleaning, filtering, mapping, or sampling data.
  - [`jq.Pipeline`](https://pkg.go.dev/go.bobheadxi.dev/streamline/jq#Pipeline) can be used to map every line to the output of a JQ query, for example.
  - [`streamline.Stream` implements standard `io` interfaces like `io.Reader`](https://pkg.go.dev/go.bobheadxi.dev/streamline#Stream.Read), so `pipeline.Pipeline` can be used for general-purpose data manipulation as well.
- [`pipe.NewStream`](https://pkg.go.dev/go.bobheadxi.dev/streamline/pipe#NewStream) offers a way to create a buffered pipe between a writer and a `Stream`.
  - [`streamexec.Start`](https://pkg.go.dev/go.bobheadxi.dev/streamline/streamexec#Start) uses this to attach a `Stream` to an `exec.Cmd` to work with command output.

When working with data streams in Go, you typically get an `io.Reader`, which is great for arbitrary data - but in many cases, especially when scripting, it's common to either end up with data and outputs that are structured line by line, or want to handle data line by line, for example to send to a structured logging library. You can set up a `bufio.Reader` or `bufio.Scanner` to do this, but for cases like `exec.Cmd` you will also need boilerplate to configure the command and set up pipes, and for additional functionality like transforming, filtering, or sampling output you will need to write your own additional handlers. `streamline` aims to provide succint ways to do all of the above and more.

### Add prefixes to command output

<table>
<tr>
  <th><code>bufio.Scanner</code></th>
  <th><code>streamline/streamexec</code></th>
</tr>
<tr>
<td>

```go
func PrefixOutput(cmd *exec.Cmd) error {
    reader, writer := io.Pipe()
    cmd.Stdout = writer
    cmd.Stderr = writer
    if err := cmd.Start(); err != nil {
        return err
    }
    errC := make(chan error)
    go func() {
        err := cmd.Wait()
        writer.Close()
        errC <- err
    }()
    s := bufio.NewScanner(reader)
    for s.Scan() {
        println("PREFIX: ", s.Text())
    }
    if err := s.Err(); err != nil {
        return err
    }
    return <-errC
}
```

</td>
<td>

```go
func PrefixOutput(cmd *exec.Cmd) error {
    stream, err := streamexec.Start(cmd)
    if err != nil {
        return err
    }
    return stream.Stream(func(line string) {
        println("PREFIX: ", line)
    })
}
```

</td>
</tr>
</table>

### Process JSON on the fly

<table>
<tr>
  <th><code>bufio.Scanner</code></th>
  <th><code>streamline</code></th>
</tr>
<tr>
<td>

```go
func GetMessages(r io.Reader) error {
    s := bufio.NewScanner(r)
    for s.Scan() {
        var result bytes.Buffer
        cmd := exec.Command("jq", ".msg")
        cmd.Stdin = bytes.NewReader(s.Bytes())
        cmd.Stdout = &result
        if err := cmd.Run(); err != nil {
            return err
        }
        print(result.String())
    }
    return s.Err()
}
```

</td>

<td>

```go
func GetMessages(r io.Reader) error {
    return streamline.New(r).
        WithPipeline(jq.Pipeline(".msg")).
        Stream(func(line string) {
            println(line)
        })
}
```

</td>
</tr>
</table>

### Sample noisy output

<table>
<tr>
  <th><code>bufio.Scanner</code></th>
  <th><code>streamline</code></th>
</tr>
<tr>
<td>

```go
func PrintEvery10th(r io.Reader) error {
    s := bufio.NewScanner(r)
    var count int
    for s.Scan() {
        count++
        if count%10 != 0 {
            continue
        }
        println(s.Text())
    }
    return s.Err()
}
```

</td>

<td>

```go
func PrintEvery10th(r io.Reader) error {
    return streamline.New(r).
        WithPipeline(pipeline.Sample(10)).
        Stream(func(line string) {
            println(line)
        })
}
```

</td>
</tr>
</table>

### Transform specific lines

This particular example is a somewhat realistic one - [GCP Cloud SQL cannot accept `pgdump` output that contains certain `EXTENSION`-related statements](https://cloud.google.com/sql/docs/postgres/import-export/import-export-dmp#external-server), so to `pgdump` a PostgreSQL database and upload the dump in a bucket for import into Cloud SQL, one must pre-process their dumps to remove offending statements.

<table>
<tr>
  <th><code>bufio.Scanner</code></th>
  <th><code>streamline</code></th>
</tr>
<tr>
<td>

```go
var unwanted = []byte("COMMENT ON EXTENSION")

func Upload(pgdump *os.File, dst io.Writer) error {
    s := bufio.NewScanner(pgdump)
    for s.Scan() {
        line := s.Bytes()
        var err error
        if bytes.Contains(line, unwanted) {
            _, err = dst.Write(
                // comment out this line
                append([]byte("-- "), line...))
        } else {
            _, err = dst.Write(line)
        }
        if err != nil {
            return err
        }
    }
    return s.Err()
}
```

</td>

<td>

```go
var unwanted = []byte("COMMENT ON EXTENSION")

func Upload(pgdump *os.File, dst io.Writer) error {
    _, err := streamline.New(pgdump).
        WithPipeline(pipeline.Map(func(line []byte) []byte {
            if bytes.Contains(line, unwanted) {
                // comment out this line
                return append([]byte("-- "), line...)
            }
            return line
        })).
        WriteTo(dst)
    return err
}
```

</td>
</tr>
</table>

## Background

Some of the ideas in this package started in [`sourcegraph/run`](https://github.com/sourcegraph/run), which started as a project trying to build utilities that [made it easier to write bash-esque scripts using Go](https://github.com/sourcegraph/sourcegraph/blob/main/doc/dev/adr/1652433602-use-go-for-scripting.md) - namely being able to do things you would often to in scripts such as grepping and iterating over lines. `streamline` generalizes on the ideas used in `sourcegraph/run` for working with command output to work on arbitrary inputs, and `sourcegraph/run` now uses `streamline` internally.
