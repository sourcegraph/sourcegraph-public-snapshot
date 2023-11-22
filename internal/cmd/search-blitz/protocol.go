package main

import "time"

type metrics struct {
	took        time.Duration
	firstResult time.Duration
	matchCount  int
	trace       string
}
