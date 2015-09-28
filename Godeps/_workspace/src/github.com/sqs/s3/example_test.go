package s3_test

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sqs/s3"
)

func ExampleClient() {
	client := s3.Client(s3.Keys{
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
	})
	resp, err := client.Get("https://example.s3.amazonaws.com/cat.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}

func ExampleClient_post() {
	client := s3.Client(s3.Keys{
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
	})
	data := strings.NewReader("hello, world")
	resp, err := client.Post("https://example.s3.amazonaws.com/foo", "text/plain", data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.StatusCode)
}

func ExampleClient_put() {
	client := s3.Client(s3.Keys{
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
	})
	data := strings.NewReader("hello, world")
	r, _ := http.NewRequest("PUT", "https://example.s3.amazonaws.com/foo", data)
	r.Header.Set("X-Amz-Acl", "public-read")
	resp, err := client.Do(r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.StatusCode)
}

func ExampleSign() {
	keys := s3.Keys{
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
	}
	data := strings.NewReader("hello, world")
	r, _ := http.NewRequest("PUT", "https://example.s3.amazonaws.com/foo", data)
	r.ContentLength = int64(data.Len())
	r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	r.Header.Set("X-Amz-Acl", "public-read")
	s3.Sign(r, keys)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.StatusCode)
}
