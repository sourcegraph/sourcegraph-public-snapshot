// Command fakehub serves git repositories within some directory over HTTP,
// along with a pastable config for easier manual testing of sourcegraph.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"testing"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func Test_fakehub(t *testing.T) {
	t.Run("empty case", func(t *testing.T) {
		// Start server.
		d, err := ioutil.TempDir("", "fakehub_test")
		if err != nil {
			t.Fatal(err)
		}
		ln, err := net.Listen("tcp", "127.0.0.1:")
		if err != nil {
			t.Fatal(err)
		}
		s, err := fakehub(1, ln, d)
		if err != nil {
			t.Fatal(err)
		}
		var eg errgroup.Group
		eg.Go(func() error {
			return s.Serve(ln)
		})

		// Main page should show config.
		addr := ln.Addr()
		confStr, err := fetch(addr, "/")
		if err != nil {
			t.Fatal(err)
		}
		type Conf struct {
			Url   string
			Repos []string
		}
		comments, err := regexp.Compile(`^//.*`)
		if err != nil {
			t.Fatal(err)
		}
		confStr = comments.ReplaceAllString(confStr, "")
		var conf Conf
		if err := json.Unmarshal([]byte(confStr), &conf); err != nil {
			t.Fatal(err, " while parsing \n", confStr)
		}

		// Clean up.
		if err := s.Shutdown(context.Background()); err != nil {
			t.Fatal(err)
		}
		if err := eg.Wait(); err != nil {
			if err != http.ErrServerClosed {
				t.Fatal(err)
			}
		}
	})
}

func fetch(addr net.Addr, path string) (string, error) {
	u := fmt.Sprint("http://", addr, path)
	resp, err := http.Get(u)
	if err != nil {
		return "", errors.Wrapf(err, "getting %s", u)
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "reading response body")
	}
	return string(contents), nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_559(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
