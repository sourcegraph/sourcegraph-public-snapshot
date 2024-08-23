# 🏃‍♂️ run

[![Go Reference](https://pkg.go.dev/badge/github.com/sourcegraph/run.svg)](https://pkg.go.dev/github.com/sourcegraph/run) ![Tests](https://github.com/sourcegraph/run/actions/workflows/pipeline.yml/badge.svg?branch=main) [![Coverage](https://codecov.io/gh/sourcegraph/run/branch/main/graph/badge.svg?token=6ELDASP2U4)](https://codecov.io/gh/sourcegraph/run)

A new way to execute commands in Go

## Example usage

<!-- START EXAMPLE -->

```go
package main

import (
  "bytes"
  "context"
  "fmt"
  "io"
  "log"
  "os"

  "github.com/sourcegraph/run"
)

func main() {
  ctx := context.Background()

  // Easily stream all output back to standard out
  err := run.Cmd(ctx, "echo", "hello world").Run().Stream(os.Stdout)
  if err != nil {
    log.Fatal(err.Error())
  }

  // Or collect, map, and modify output, then collect string lines from it
  lines, err := run.Cmd(ctx, "ls").Run().
    Map(func(ctx context.Context, line []byte, dst io.Writer) (int, error) {
      if !bytes.HasSuffix(line, []byte(".go")) {
        return 0, nil
      }
      return dst.Write(bytes.TrimSuffix(line, []byte(".go")))
    }).
    Lines()
  if err != nil {
    log.Fatal(err.Error())
  }
  for i, l := range lines {
    fmt.Printf("line %d: %q\n", i, l)
  }

  // Errors include standard error by default, so we can just stream stdout.
  err = run.Cmd(ctx, "ls", "foobar").StdOut().Run().Stream(os.Stdout)
  if err != nil {
    println(err.Error()) // exit status 1: ls: foobar: No such file or directory
  }

  // Generate data from a file, replacing tabs with spaces for Markdown purposes
  var exampleData bytes.Buffer
  exampleData.Write([]byte(exampleStart + "\n\n```go\n"))
  if err = run.Cmd(ctx, "cat", "cmd/example/main.go").Run().
    Map(func(ctx context.Context, line []byte, dst io.Writer) (int, error) {
      return dst.Write(bytes.ReplaceAll(line, []byte("\t"), []byte("  ")))
    }).
    Stream(&exampleData); err != nil {
    log.Fatal(err)
  }
  exampleData.Write([]byte("```\n\n" + exampleEnd))

  // Render new README file
  var readmeData bytes.Buffer
  if err = run.Cmd(ctx, "cat", "README.md").Run().Stream(&readmeData); err != nil {
    log.Fatal(err)
  }
  replaced := exampleBlockRegexp.ReplaceAll(readmeData.Bytes(), exampleData.Bytes())

  // Pipe data to command
  err = run.Cmd(ctx, "cp /dev/stdin README.md").Input(bytes.NewReader(replaced)).Run().Wait()
  if err != nil {
    log.Fatal(err)
  }
}
```

<!-- END EXAMPLE -->
