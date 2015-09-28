// Command s3cp copies a file to or from Amazon S3.
//
// Usage:
//
//   s3cp file url
//   s3cp url file
//
// The file does not need to be seekable or stat-able. You can use s3cp to
// upload data of indeterminate length, such as from a pipe.
//
// Examples:
//   $ s3cp file.txt https://mybucket.s3.amazonaws.com/file.txt
//   $ gendata | s3cp /dev/stdin https://mybucket.s3.amazonaws.com/log
//   $ s3cp https://mybucket.s3.amazonaws.com/image.jpg pic.jpg
//
// Environment:
//
// S3_ACCESS_KEY – an AWS Access Key Id (required)
//
// S3_SECRET_KEY – an AWS Secret Access Key (required)
package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sqs/s3/s3util"
)

func main() {
	s3util.DefaultConfig.AccessKey = os.Getenv("S3_ACCESS_KEY")
	s3util.DefaultConfig.SecretKey = os.Getenv("S3_SECRET_KEY")
	args := os.Args[1:]
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: s3cp file url")
		fmt.Fprintln(os.Stderr, "       s3cp url file")
		os.Exit(1)
	}

	r, err := open(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	w, err := create(args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, err = io.Copy(w, r)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = w.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func open(s string) (io.ReadCloser, error) {
	if isURL(s) {
		return s3util.Open(s, nil)
	}
	return os.Open(s)
}

func create(s string) (io.WriteCloser, error) {
	if isURL(s) {
		return s3util.Create(s, nil, nil)
	}
	return os.Create(s)
}

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
