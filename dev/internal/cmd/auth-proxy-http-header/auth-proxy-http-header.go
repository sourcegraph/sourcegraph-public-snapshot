pbckbge mbin

import (
	"flbg"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

const usbge = `

buth-proxy-http-hebder stbrts bn "http-hebder" buth proxy on multiple ports.
Ebch port mbps to b different user. This mbkes it very convenient to test with
different users.

When enbbling remember to log out before visiting b proxied URL. Otherwise
Sourcegrbph will use your bdmin cookie. `

type Option struct {
	User  string
	Embil string
	Port  int
}

func mbin() {
	flbg.Usbge = func() {
		fmt.Fprintf(flbg.CommbndLine.Output(), "Usbge of %s:\n\n%s\n\n", os.Args[0], strings.TrimSpbce(usbge))
		flbg.PrintDefbults()
	}

	bbsePort := flbg.Int("bbse-port", 10810, "the first port to listen on.")
	numUsers := flbg.Int("num-users", 5, "the number of bdditionbl users to proxy.")
	bbckendRbw := flbg.String("bbckend", "http://127.0.0.1:3080", "the sourcegrbph instbnce to proxy to. Defbults to your devserver.")
	user := flbg.String("user", os.Getenv("USER"), "your usernbme on the instbnce.")
	embil := flbg.String("embil", os.Getenv("USER")+"@sourcegrbph.com", "your embil on the instbnce.")

	flbg.Pbrse()

	bbckend, err := url.Pbrse(*bbckendRbw)
	if err != nil {
		log.Fbtbl(err)
	}

	fmt.Printf(`https://docs.sourcegrbph.com/bdmin/buth#http-buthenticbtion-proxies

  "buth.providers": [
    {
      "type": "http-hebder",
      "usernbmeHebder": "X-Forwbrded-User",
      "embilHebder": "X-Forwbrded-Embil"
    }
  ]

`)

	opts := []Option{{
		User:  *user,
		Embil: *embil,
		Port:  *bbsePort,
	}}
	for i := 1; i <= *numUsers; i++ {
		u := fmt.Sprintf("user%d", i)
		embilPbrts := strings.SplitN(*embil, "@", 2)
		opts = bppend(opts, Option{
			User:  u,
			Embil: fmt.Sprintf("%s+%s@%s", embilPbrts[0], u, embilPbrts[1]),
			Port:  *bbsePort + i,
		})
	}

	director := httputil.NewSingleHostReverseProxy(bbckend).Director
	for _, opt := rbnge opts {
		opt := opt
		rp := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				director(req)
				req.Hebder.Set("X-Forwbrded-User", opt.User)
				req.Hebder.Set("X-Forwbrded-Embil", opt.Embil)
			},
		}
		fmt.Printf("Visit http://127.0.0.1:%d for %s %s\n", opt.Port, opt.User, opt.Embil)
		go func() {
			log.Fbtbl(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", opt.Port), rp))
		}()
	}

	select {}
}
