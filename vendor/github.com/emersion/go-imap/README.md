# go-imap

[![GoDoc](https://godoc.org/github.com/emersion/go-imap?status.svg)](https://godoc.org/github.com/emersion/go-imap)
[![Build Status](https://travis-ci.org/emersion/go-imap.svg?branch=master)](https://travis-ci.org/emersion/go-imap)
[![Codecov](https://codecov.io/gh/emersion/go-imap/branch/master/graph/badge.svg)](https://codecov.io/gh/emersion/go-imap)
[![Go Report
Card](https://goreportcard.com/badge/github.com/emersion/go-imap)](https://goreportcard.com/report/github.com/emersion/go-imap)
[![Unstable](https://img.shields.io/badge/stability-unstable-yellow.svg)](https://github.com/emersion/stability-badges#unstable)
[![Gitter chat](https://badges.gitter.im/goimap/Lobby.svg)](https://gitter.im/goimap/Lobby)

An [IMAP4rev1](https://tools.ietf.org/html/rfc3501) library written in Go. It
can be used to build a client and/or a server.

```bash
go get github.com/emersion/go-imap/...
```

## Why?

Other IMAP implementations in Go:
* Require to make [many type assertions or conversions](https://github.com/emersion/neutron/blob/ca635850e2223d6cfe818664ef901fa6e3c1d859/backend/imap/util.go#L110)
* Are not idiomatic or are [ugly](https://github.com/jordwest/imap-server/blob/master/conn/commands.go#L53)
* Are [not pleasant to use](https://github.com/emersion/neutron/blob/ca635850e2223d6cfe818664ef901fa6e3c1d859/backend/imap/messages.go#L228)
* Implement a server _xor_ a client, not both
* Don't implement unilateral updates (i.e. the server can't notify clients for
  new messages)
* Do not have a good test coverage
* Don't handle encoding and charset automatically

## Usage

### Client [![GoDoc](https://godoc.org/github.com/emersion/go-imap/client?status.svg)](https://godoc.org/github.com/emersion/go-imap/client)

```go
package main

import (
	"log"

	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap"
)

func main() {
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS("mail.example.org:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login("username", "password"); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func () {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("Mailboxes:")
	for m := range mailboxes {
		log.Println("* " + m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Flags for INBOX:", mbox.Flags)

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 3 {
		// We're using unsigned integers here, only substract if the result is > 0
		from = mbox.Messages - 3
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	log.Println("Last 4 messages:")
	for msg := range messages {
		log.Println("* " + msg.Envelope.Subject)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}
```

### Server [![GoDoc](https://godoc.org/github.com/emersion/go-imap/server?status.svg)](https://godoc.org/github.com/emersion/go-imap/server)

```go
package main

import (
	"log"

	"github.com/emersion/go-imap/server"
	"github.com/emersion/go-imap/backend/memory"
)

func main() {
	// Create a memory backend
	be := memory.New()

	// Create a new server
	s := server.New(be)
	s.Addr = ":1143"
	// Since we will use this server for testing only, we can allow plain text
	// authentication over unencrypted connections
	s.AllowInsecureAuth = true

	log.Println("Starting IMAP server at localhost:1143")
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
```

You can now use `telnet localhost 1143` to manually connect to the server.

## Extending go-imap

### Extensions

Commands defined in IMAP extensions are available in other packages. See [the
wiki](https://github.com/emersion/go-imap/wiki/Using-extensions#using-client-extensions)
to learn how to use them.

* [APPENDLIMIT](https://github.com/emersion/go-imap-appendlimit)
* [COMPRESS](https://github.com/emersion/go-imap-compress)
* [ENABLE](https://github.com/emersion/go-imap-enable)
* [ID](https://github.com/ProtonMail/go-imap-id)
* [IDLE](https://github.com/emersion/go-imap-idle)
* [MOVE](https://github.com/emersion/go-imap-move)
* [QUOTA](https://github.com/emersion/go-imap-quota)
* [SPECIAL-USE](https://github.com/emersion/go-imap-specialuse)
* [UNSELECT](https://github.com/emersion/go-imap-unselect)
* [UIDPLUS](https://github.com/emersion/go-imap-uidplus)

### Server backends

* [Memory](https://github.com/emersion/go-imap/tree/master/backend/memory) (for testing)
* [Multi](https://github.com/emersion/go-imap-multi)
* [PGP](https://github.com/emersion/go-imap-pgp)
* [Proxy](https://github.com/emersion/go-imap-proxy)

### Related projects

* [go-message](https://github.com/emersion/go-message) - parsing and formatting MIME and mail messages
* [go-pgpmail](https://github.com/emersion/go-pgpmail) - decrypting and encrypting mails with OpenPGP
* [go-sasl](https://github.com/emersion/go-sasl) - sending and receiving SASL authentications
* [go-smtp](https://github.com/emersion/go-smtp) - building SMTP clients and servers
* [go-dkim](https://github.com/emersion/go-dkim) - creating and verifying DKIM signatures

## License

MIT
