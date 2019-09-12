package main

import (
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestInstrumentHandler(t *testing.T) {
	h := http.Handler(nil)
	instrumentHandler(prometheus.DefaultRegisterer, h)
}
