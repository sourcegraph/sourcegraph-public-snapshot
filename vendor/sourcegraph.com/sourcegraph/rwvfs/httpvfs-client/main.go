package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"time"

	"golang.org/x/tools/godoc/vfs"

	"strings"

	"sourcegraph.com/sourcegraph/rwvfs"
)

var (
	urlStr    = flag.String("url", "http://localhost:7070/", "URL to HTTP VFS (typically served by the `httpvfs` program)")
	startByte = flag.Int("start-byte", 0, "('cat' only) byte range start")
	endByte   = flag.Int("end-byte", -1, "('cat' only) byte range end")
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	if flag.NArg() != 2 {
		log.Fatal("error: usage: httpvfs-client [opts] <cat|ls|put|rm> <path>")
	}
	op := flag.Arg(0)
	path := path.Clean(flag.Arg(1))

	url, err := url.Parse(*urlStr)
	if err != nil {
		log.Fatal(err)
	}

	fs := rwvfs.HTTP(url, nil)

	switch strings.ToLower(op) {
	case "cat":
		if *startByte != 0 && *endByte < *startByte {
			log.Fatal("error: -end-byte must be greater than -start-byte")
		}

		var f vfs.ReadSeekCloser
		var err error

		if *startByte != 0 {
			f, err = fs.(rwvfs.FetcherOpener).OpenFetcher(path)
		} else {
			f, err = fs.Open(path)
		}
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		rdr := io.Reader(f)

		if *startByte != 0 {
			if err := f.(rwvfs.Fetcher).Fetch(int64(*startByte), int64(*endByte)); err != nil {
				log.Fatalf("Fetch bytes=%d-%d: %s", *startByte, *endByte, err)
			}
		}

		if *startByte != 0 {
			if _, err := f.Seek(int64(*startByte), 0); err != nil {
				log.Fatalln("Seek:", err)
			}
		}
		if *endByte != -1 {
			byteLen := *endByte - *startByte
			rdr = io.LimitReader(f, int64(byteLen))
		}

		if _, err := io.Copy(os.Stdout, rdr); err != nil {
			log.Fatalln("Copy:", err)
		}

	case "ls":
		fis, err := fs.ReadDir(path)
		if err != nil {
			log.Fatal(err)
		}

		var longestNameLen int
		for _, fi := range fis {
			if len(fi.Name()) > longestNameLen {
				longestNameLen = len(fi.Name())
			}
		}
		longestNameLen++ // account for "/" suffix on dirs

		for _, fi := range fis {
			name := fi.Name()
			if fi.IsDir() {
				name += "/"
			}
			mtime := fi.ModTime().Round(time.Second)
			fmt.Printf("%-*s   %s   %d\n", longestNameLen, name, mtime, fi.Size())
		}

	case "put":
		f, err := fs.Create(path)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		}()
		log.Println("(reading file data on stdin...)")
		if _, err := io.Copy(f, os.Stdin); err != nil {
			log.Fatal(err)
		}

	case "rm":
		if err := fs.Remove(path); err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatal("error: invalid op (see -h)")
	}
}
