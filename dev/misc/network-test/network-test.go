// The network-test command tests dialing and listening on addrs given
// in the flags.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var (
	dialAddr   = flag.String("dial", "", "TCP addr to dial, such as 'localhost:12345' (another process should be listening on this addr)")
	clients    = flag.Int("clients", 50, "number of concurrent client connections to the dial addr")
	listenAddr = flag.String("listen", "", "TCP addr to listen on, such as ':12345' (use ':0' to choose an available high-numbered port)")
)

func main() {
	log.SetFlags(0)
	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(1)
	}
	if *dialAddr == "" && *listenAddr == "" {
		log.Println(`The network-test command tests dialing and listening on TCP addresses given in the flags -dial and -listen (run with -h for more information).

To run a simple test, execute the following commands in different terminals (so they can run simultaneously) on the same machine:

$ ./network-test -listen :12345
$ ./network-test -dial localhost:12345

You can replace 12345 with any other port number. The client (the command with -dial) attempts to make numerous concurrent connections to localhost:12345 and reports if any of them fail.
`)
	}
	if *dialAddr != "" && *listenAddr != "" {
		log.Fatal("Exactly one of -dial and -listen must be specified. Run two separate processes to test both ends of the connection.")
	}
	if *dialAddr != "" {
		var wg sync.WaitGroup
		for i := 0; i < *clients; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				dialTest(*dialAddr)
			}()
		}
		wg.Wait()
	}
	if *listenAddr != "" {
		listenTest(*listenAddr)
	}
}

const (
	numPings = 3
	wait     = time.Millisecond * 200
)

func dialTest(addr string) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf(`Dial("tcp", %q) error: %s.`, addr, err)
	}
	log.Printf("Dialed %s OK.", connInfo(c))
	for i := 0; i < numPings; i++ {
		n, err := c.Write([]byte("ping"))
		if err != nil {
			log.Fatalf("Write #%d on %s failed: %s (n=%d).", i+1, connInfo(c), err, n)
		}

		pong := make([]byte, 4)
		n, err = c.Read(pong)
		if err != nil {
			log.Fatalf("Read #%d on %s failed: %s (n=%d).", i+1, connInfo(c), err, n)
		}
		if !bytes.Equal(pong, []byte("pong")) {
			log.Printf(`Received server message %q, wanted "pong".`, pong)
		}

		time.Sleep(wait)
	}
	if err := c.Close(); err != nil {
		log.Fatalf("Close %s error: %s.", connInfo(c), err)
	}
	log.Printf("Closed %s OK.", connInfo(c))
}

func listenTest(addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf(`Listen("tcp", %q) error: %s.`, addr, err)
	}
	log.Printf("Listened on %s OK.", l.Addr())
	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatalf("Accept error: %s.", err)
		}
		log.Printf("Accepted client connection %s OK.", connInfo(c))

		go func() {
			for i := 0; i < numPings; i++ {
				ping := make([]byte, 4)
				n, err := c.Read(ping)
				if err != nil {
					log.Fatalf("Read #%d on %s failed: %s (n=%d).", i+1, connInfo(c), err, n)
				}
				if !bytes.Equal(ping, []byte("ping")) {
					log.Printf(`Received client message %q, wanted "ping".`, ping)
				}

				n, err = c.Write([]byte("pong"))
				if err != nil {
					log.Fatalf("Write #%d on %s failed: %s (n=%d).", i+1, connInfo(c), err, n)
				}
			}
			if err := c.Close(); err != nil {
				log.Fatalf("Close %s error: %s.", connInfo(c), err)
			}
			log.Printf("Closed %s OK.", connInfo(c))
		}()
	}
}

func connInfo(c net.Conn) string {
	return fmt.Sprintf("%s<->%s", c.LocalAddr(), c.RemoteAddr())
}
