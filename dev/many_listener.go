// +build ignore

// The many_listener command listens on all ports in the specified
// range. It allows you to simulate being on a host where many ports
// are already in use.
package main

import (
	"flag"
	"log"
	"math"
	"net"
	"strconv"
	"sync"
)

var (
	loPort = flag.Int("lo", 10000, "beginning of port range to listen on (inclusive)")
	hiPort = flag.Int("hi", 10010, "end of port range to listen on (inclusive)")
)

func checkPort(n int) {
	if n <= 0 || n > math.MaxUint16 {
		log.Fatalf("invalid port number %d", n)
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	var wg sync.WaitGroup
	for port := *loPort; port <= *hiPort; port++ {
		wg.Add(1)
		go func(port int) {
			if err := listen(port, &wg); err != nil {
				log.Printf("port %d listener: %s", port, err)
			}
		}(port)
	}

	wg.Wait()
	log.Printf("listening on port range %d-%d (inclusive)", *loPort, *hiPort)

	select {}
}

func listen(port int, wg *sync.WaitGroup) error {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	wg.Done()

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		log.Printf("%d", port)
		if err := conn.Close(); err != nil {
			return err
		}
	}
}
