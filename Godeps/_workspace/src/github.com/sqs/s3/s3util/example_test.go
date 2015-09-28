package s3util_test

import (
	"fmt"
	"io"
	"os"

	"github.com/sqs/s3/s3util"
)

func ExampleCreate() {
	s3util.DefaultConfig.AccessKey = "...access key..."
	s3util.DefaultConfig.SecretKey = "...secret key..."
	r, _ := os.Open("/dev/stdin")
	w, _ := s3util.Create("https://mybucket.s3.amazonaws.com/log.txt", nil, nil)
	io.Copy(w, r)
	w.Close()
}

func ExampleDelete() {
	s3util.DefaultConfig.AccessKey = "...access key..."
	s3util.DefaultConfig.SecretKey = "...secret key..."
	r, _ := s3util.Delete("https://mybucket.s3.amazonaws.com/log.txt", nil)
	r.Close()
}

func ExampleOpen() {
	s3util.DefaultConfig.AccessKey = "...access key..."
	s3util.DefaultConfig.SecretKey = "...secret key..."
	r, _ := s3util.Open("https://mybucket.s3.amazonaws.com/log.txt", nil)
	w, _ := os.Create("out.txt")
	io.Copy(w, r)
	w.Close()
}

func ExampleReaddir() {
	s3util.DefaultConfig.AccessKey = os.Getenv("S3_ACCESS_KEY")
	s3util.DefaultConfig.SecretKey = os.Getenv("S3_SECRET_KEY")
	f, err := s3util.NewFile("https://examle.s3.amazonaws.com/foo", nil)
	if err != nil {
		panic(err)
	}
	var infos []os.FileInfo
	for {
		infos, err = f.Readdir(0)
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		for i, info := range infos {
			c := info.Sys().(*s3util.Stat)
			var etag string
			if c != nil {
				etag = c.ETag
			}
			fmt.Printf("%d: %v, %s\n", i, info, etag)
		}
	}
}
