package main

import (
	"context"
	"io"
	"net"
	"net/http"

	"github.com/mxk/go-flowrate/flowrate"
)

type connReadWriter struct {
	net.Conn

	Reader io.Reader
	Writer io.Writer
}

func (c *connReadWriter) Read(b []byte) (int, error) {
	return c.Reader.Read(b)
}

func (c *connReadWriter) Write(b []byte) (int, error) {
	return c.Writer.Write(b)
}

type dial func(ctx context.Context, network, addr string) (net.Conn, error)

func limitDial(d dial, limit int64) dial {
	if limit <= 0 {
		return d
	}

	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := d(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		return &connReadWriter{
			Conn:   conn,
			Reader: flowrate.NewReader(conn, limit),
			Writer: flowrate.NewWriter(conn, limit),
		}, nil
	}
}

func limitHTTPDefaultClient(limitMbps int64) {
	if limitMbps <= 0 {
		return
	}

	const megabit = 1000 * 1000
	limit := (limitMbps * megabit) / 8

	t := http.DefaultTransport.(*http.Transport)
	t.DialContext = limitDial(t.DialContext, limit)
}
