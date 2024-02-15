package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

const usage = `

auth-proxy-http-header starts an "http-header" auth proxy on multiple ports.
Each port maps to a different user. This makes it very convenient to test with
different users.

When enabling remember to log out before visiting a proxied URL. Otherwise
Sourcegraph will use your admin cookie. `

type Option struct {
	User  string
	Email string
	Port  int
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n\n%s\n\n", os.Args[0], strings.TrimSpace(usage))
		flag.PrintDefaults()
	}

	basePort := flag.Int("base-port", 10810, "the first port to listen on.")
	numUsers := flag.Int("num-users", 5, "the number of additional users to proxy.")
	backendRaw := flag.String("backend", "http://127.0.0.1:3080", "the sourcegraph instance to proxy to. Defaults to your devserver.")
	user := flag.String("user", os.Getenv("USER"), "your username on the instance.")
	email := flag.String("email", os.Getenv("USER")+"@sourcegraph.com", "your email on the instance.")

	flag.Parse()

	backend, err := url.Parse(*backendRaw)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(`https://sourcegraph.com/docs/admin/auth#http-authentication-proxies

  "auth.providers": [
    {
      "type": "http-header",
      "usernameHeader": "X-Forwarded-User",
      "emailHeader": "X-Forwarded-Email"
    }
  ]

`)

	opts := []Option{{
		User:  *user,
		Email: *email,
		Port:  *basePort,
	}}
	for i := 1; i <= *numUsers; i++ {
		u := fmt.Sprintf("user%d", i)
		emailParts := strings.SplitN(*email, "@", 2)
		opts = append(opts, Option{
			User:  u,
			Email: fmt.Sprintf("%s+%s@%s", emailParts[0], u, emailParts[1]),
			Port:  *basePort + i,
		})
	}

	director := httputil.NewSingleHostReverseProxy(backend).Director
	for _, opt := range opts {
		opt := opt
		rp := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				director(req)
				req.Header.Set("X-Forwarded-User", opt.User)
				req.Header.Set("X-Forwarded-Email", opt.Email)
			},
		}
		fmt.Printf("Visit http://127.0.0.1:%d for %s %s\n", opt.Port, opt.User, opt.Email)
		go func() {
			log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", opt.Port), rp))
		}()
	}

	select {}
}
