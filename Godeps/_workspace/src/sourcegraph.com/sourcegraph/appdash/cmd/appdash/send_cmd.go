package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"sourcegraph.com/sourcegraph/appdash"
)

func init() {
	_, err := CLI.AddCommand("send",
		"send sample data to a remote collector",
		"The send command sends sample data to a remote collector server.",
		&sendCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
}

// SendCmd is the command for running Appdash in sender mode, where it sends
// sample data to a remote collector.
type SendCmd struct {
	CollectorAddr  string `short:"c" long:"collector" description:"collector listen address" default:":7701"`
	CollectorProto string `short:"p" long:"proto" description:"collector protocol (tcp or tls)" default:"tcp"`
	ServerName     string `short:"s" long:"server-name" description:"server name (required for TLS)"`
	Debug          bool   `short:"d" long:"debug" description:"debug log"`
}

var sendCmd SendCmd

// Execute execudes the commands with the given arguments and returns an error,
// if any.
func (c *SendCmd) Execute(args []string) error {
	var rc *appdash.RemoteCollector
	switch c.CollectorProto {
	case "tcp":
		rc = appdash.NewRemoteCollector(c.CollectorAddr)
	case "tls":
		rc = appdash.NewTLSRemoteCollector(c.CollectorAddr, &tls.Config{ServerName: c.ServerName})
	default:
		return fmt.Errorf("unknown proto: %q", c.CollectorProto)
	}
	rc.Debug = c.Debug

	rcc := &appdash.ChunkedCollector{
		Collector:   rc,
		MinInterval: time.Second,
	}

	log.Println("Sending sample data...")
	if err := sampleData(rc); err != nil {
		return err
	}
	log.Println("Done sending sample data.")

	log.Println("Flushing chunked collector...")
	if err := rcc.Flush(); err != nil {
		return err
	}
	log.Println("Done flushing chunked collector.")

	return nil
}
