package lightstep

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/opentracing/basictracer-go"
	"github.com/opentracing/opentracing-go"
)

const (
	nanosPerMicro = 1000

	assembleTraceAPIPath = "/api/v1/trace/assemble"

	// HTTP header for conveying the access token for web APIs
	accessTokenHeader = "LightStep-Access-Token"
)

var (
	ErrNotLightStepTracer = errors.New("not a LightStep tracer")
	ErrSpanIsTooOld       = errors.New("span is too old to assemble")
)

type httpError string

// AssembleTraceForSpan requests trace assembly given a span of
// interest to the caller.  The span may not have had Finish() called.
func AssembleTraceForSpan(span basictracer.Span) error {
	return assembleTraceBy(span, func(span basictracer.Span) []byte {
		// Note: API handler expects span_guid to be a string,
		// for consistency with other handlers.
		return []byte(fmt.Sprint(`{"span_guid":"`, span.Context().(*basictracer.SpanContext).SpanID,
			`","at_micros": `, time.Now().UnixNano()/nanosPerMicro, `}`))
	})
}

func assembleTraceBy(span opentracing.Span, payload func(span basictracer.Span) []byte) error {
	bspan, ok := span.(basictracer.Span)
	if !ok {
		return ErrNotLightStepTracer
	}
	btracer, ok := span.Tracer().(basictracer.Tracer)
	if !ok {
		return ErrNotLightStepTracer
	}
	recorder, ok := btracer.Options().Recorder.(*Recorder)
	if !ok {
		return ErrNotLightStepTracer
	}

	url := recorder.apiURL + assembleTraceAPIPath
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload(bspan)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(accessTokenHeader, recorder.accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return httpError(resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var rpayload struct {
		AssemblyStatus string `json:"assembly_status"`
	}
	if err := json.Unmarshal(body, &rpayload); err != nil {
		return err
	}
	if rpayload.AssemblyStatus == "lost" {
		return ErrSpanIsTooOld
	}
	return nil
}

func (h httpError) Error() string {
	return string(h)
}
