package main

import "golang.org/x/net/websocket"

var (
	origin = "http://localhost/"
	addr   = "ws://localhost:4100/pipeline"
	codec  = websocket.JSON
)
