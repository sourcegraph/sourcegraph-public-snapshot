package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strings"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/base"
	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/base/imath"
	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument"
	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument/client"
)

func spazzOut(numSpanClasses int, sleepMean, sleepStddev base.Micros, payloadFieldsMean, payloadFieldsStddev, messageLenMean, messageLenStddev int) {
	for {
		activeSpan := instrument.StartSpan()
		activeSpan.SetOperation(fmt.Sprintf("testing/op_%v", rand.Intn(numSpanClasses)))
		numPayloadFields := payloadFieldsMean + int(rand.NormFloat64()*float64(payloadFieldsStddev))
		payload := make(map[int]string)
		for i := 0; i < numPayloadFields; i++ {
			payload[i] = fmt.Sprintf("Payload field #%v", i)
		}

		messageLen := imath.Max(1, messageLenMean+int(rand.NormFloat64()*float64(messageLenStddev)))
		instrument.Log(instrument.Print(strings.Repeat("m", messageLen)).Payload(payload))

		sleepMicros := (sleepMean + base.Micros(rand.NormFloat64()*float64(sleepStddev))).Max(0)
		activeSpan.Log(fmt.Sprintf("sleeping for %v micros", sleepMicros))
		time.Sleep(time.Duration(sleepMicros) * time.Microsecond)
		activeSpan.Finish()
	}
}

func main() {
	runtime.MemProfileRate = 512
	flag.Parse()
	instrument.SetDefaultRuntime(client.NewRuntime(
		&client.Options{
			AccessToken: "invalid",
			ServiceHost: "localhost",
		}))

	go spazzOut(100, 5000, 1000, 250, 150, 100000, 50000)

	go func() {
		for _ = range time.Tick(2 * time.Second) {
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			fmt.Printf(`Memory stats:
	HeapAlloc: %vMB
	HeapSys:   %vMB
`, ms.HeapAlloc/(1024*1024), ms.HeapSys/(1024*1024))
		}
	}()

	// For pprof.
	go http.ListenAndServe(":4000", nil)

	runtime.Goexit()
}
