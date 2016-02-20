// Command appdash runs the Appdash web UI from the command-line.
//
// For information about Appdash see:
//
// https://sourcegraph.com/sourcegraph/appdash
//
// Demo mode
//
// A demo of Appdash in a small web application can be ran by simply running:
//
//  appdash demo
//
// Which will produce some output:
//
//  Appdash collector listening on tcp:46346
//  Appdash web UI running at http://localhost:8700
//
//  Appdash demo app running at http://localhost:8699
//
// Visiting the demo app URL mentioned above will then bring up the demo app
// which will make a few fake API calls, and then give you a direct link to
// view the trace for your request in Appdash's web UI.
//
// Serve mode
//
// Basic usage consists of running:
//
//  appdash serve
//
// Which will start a Appdash collector server running on TCP port 7701 in
// plain-text (i.e. insecure). The Appdash collector server can then receive
// information from your application via a appdash.NewRemoteCollector, which it
// will then display in the web UI.
//
// The web UI is also ran on HTTP port 7700, which you could visit in a
// browser:
//
//  http://localhost:7700
//
// Optionally, you do not need to use this command at all and can embed the web
// UI into your application directly on a separate HTTP port (see the traceapp
// package or examples/cmd/webapp for more details).
//
// Send mode
//
// For testing purposes, the appdash command can send some fake data to a
// remote Appdash collector server by running:
//
//  appdash send -c="localhost:7701"
//
package main

import (
	"log"
	_ "net/http/pprof"
	"os"

	"github.com/jessevdk/go-flags"
)

// CLI is the go-flags CLI object that parses command-line arguments and runs commands.
var CLI = flags.NewNamedParser("appdash", flags.Default)

// GlobalOpt contains global options.
var GlobalOpt struct {
	Verbose bool `short:"v" description:"show verbose output"`
}

func init() {
	CLI.LongDescription = "appdash is an application tracing system"
	CLI.AddGroup("Global options", "", &GlobalOpt)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("")
	if _, err := CLI.Parse(); err != nil {
		os.Exit(1)
	}
}
