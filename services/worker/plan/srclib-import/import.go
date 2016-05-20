package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/jpillora/backoff"
)

const maxAttempts = 5

func main() {
	if len(os.Args) <= 2 {
		log.Fatalf("USAGE: %s IMPORT_URL ZIP_FILE", os.Args[0])
	}
	url := os.Args[1]
	zipPath := os.Args[2]
	err := importWithRetry(url, zipPath)
	if err != nil {
		log.Fatal(err)
	}
}

func importWithRetry(url, zipPath string) error {
	b := backoff.Backoff{Jitter: true}
	for i := 1; i < maxAttempts; i++ {
		err := putZip(url, zipPath)
		if err == nil {
			return nil
		}
		log.Printf("Request to %s failed (attempt %d/%d): %s", url, i, maxAttempts, err)
		time.Sleep(b.Duration())
	}
	return putZip(url, zipPath)
}

func putZip(url, zipPath string) error {
	body, err := os.Open(zipPath)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return err
	}
	discoverAndSetAuth(req)
	// We expect the file to be a zipfile
	req.Header.Set("Content-Type", "application/x-zip-compressed")
	req.Header.Set("Content-Transfer-Encoding", "binary")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if appdash := resp.Header.Get("X-Appdash-Trace"); appdash != "" {
		log.Println("X-Appdash-Trace:", appdash)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("bad status code %d", resp.StatusCode)
	}
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

func discoverAndSetAuth(req *http.Request) {
	m := getNetRC(req.URL)
	if m == nil {
		return
	}
	log.Printf("Using basicauth from netrc for %s", m.Name)
	req.SetBasicAuth(m.Login, m.Password)
}

func getNetRC(url *url.URL) *netrc.Machine {
	rc, err := netrc.ParseFile(os.ExpandEnv("$HOME/.netrc"))
	if err != nil {
		return nil
	}
	host := url.Host
	if i := strings.Index(host, ":"); i != -1 {
		host = host[:i]
	}
	return rc.FindMachine(host)
}
