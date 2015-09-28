package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/sqltrace"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// randomSplit splits the given value randomly into N segments. The approach
// used is described at:
//
// http://stackoverflow.com/questions/22380890/generate-n-random-numbers-whose-sum-is-m-and-all-numbers-should-be-greater-than
//
func randomSplit(value, n int) []int {
	var (
		segments = make([]int, 0, n)
		sum      int
	)
	for i := 1; i < n; i++ {
		s := rand.Intn(((value - sum) / (n - i)) + 1)
		segments = append(segments, s)
		sum += s
	}
	return append(segments, value-sum)
}

func sampleData(c appdash.Collector) error {
	const numTraces = 60
	log.Printf("Adding sample data (%d traces)", numTraces)
	for i := appdash.ID(1); i <= numTraces; i++ {
		traceID := appdash.NewRootSpanID()
		traceRec := appdash.NewRecorder(traceID, c)
		traceRec.Name(fakeHosts[rand.Intn(len(fakeHosts))])

		// A random length for the trace.
		length := time.Duration(rand.Intn(1000)) * time.Millisecond

		startTime := time.Now().Add(-time.Duration(rand.Intn(100)) * time.Minute)
		traceRec.Event(&sqltrace.SQLEvent{
			ClientSend: startTime,
			ClientRecv: startTime.Add(length),
			SQL:        "SELECT * FROM table_name;",
			Tag:        fmt.Sprintf("fakeTag%d", rand.Intn(10)),
		})

		// We'll split the trace into N (3-7) spans (i.e. "N operations") each with
		// a random duration of time adding up to the length of the trace.
		numSpans := rand.Intn(7-3) + 3
		times := randomSplit(int(length/time.Millisecond), numSpans)

		lastSpanID := traceID
		for j := 1; j <= numSpans; j++ {
			// The parent span is the predecessor.
			spanID := appdash.NewSpanID(lastSpanID)

			rec := appdash.NewRecorder(spanID, c)
			rec.Name(fakeNames[(j+int(i))%len(fakeNames)])
			if j%3 == 0 {
				rec.Log("hello")
			}
			if j%5 == 0 {
				rec.Msg("hi")
			}

			// Generate a span event.
			spanDuration := time.Duration(times[j-1]) * time.Millisecond
			rec.Event(&sqltrace.SQLEvent{
				ClientSend: startTime,
				ClientRecv: startTime.Add(spanDuration),
				SQL:        "SELECT * FROM table_name;",
				Tag:        fmt.Sprintf("fakeTag%d", rand.Intn(10)),
			})

			// Shift the start time forward.
			startTime = startTime.Add(spanDuration)

			// Check for any recorder errors.
			if errs := rec.Errors(); len(errs) > 0 {
				return fmt.Errorf("recorder errors: %v", errs)
			}

			lastSpanID = spanID
		}
	}
	return nil
}

var fakeNames = []string{
	"Phafsea",
	"Kraesey",
	"Bleland",
	"Moonuiburg",
	"Zriozruamwell",
	"Erento",
	"Gona",
	"Frence",
	"Hiuwront",
	"Shuplin",
	"Luoron",
	"Eproling",
	"Iwruuhville",
	"Ripherough",
	"Sekhunsea",
	"Yery",
	"Fia",
	"Jouver",
	"Strayolis",
	"Grisaso",
}

var fakeHosts = []string{
	"api.phafsea.org",
	"web.kraesey.net",
	"www3.bleland.com",
	"mun.moonuiburg",
	"zri.ozruamwe.ll",
	"e.rento",
	"go.na",
	"fre.nce",
	"hiu.wront",
	"shu.plin:9090",
	"luoron.net",
	"api.eproling.org",
	"iw.ruuh.ville",
	"riphero.ugh",
	"sek.hun.sea",
	"api.ye.ry",
	"fia.com",
	"jouver.io",
	"strayolis.io",
	"grisaso.io",
}
