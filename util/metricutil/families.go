package metricutil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"

	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// MetricFamilies represets a collection of distinct metrics. This is a
// concept use heavily in prometheus internally
type MetricFamilies map[string]*dto.MetricFamily

// UnmarshalMetricFamilies unmarshals from the provided reader
func UnmarshalMetricFamilies(r io.Reader) MetricFamilies {
	mfs := MetricFamilies{}
	for {
		mf := &dto.MetricFamily{}
		if _, err := pbutil.ReadDelimited(r, mf); err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		mfs[mf.GetName()] = mf
	}
	return mfs
}

// SnapshotMetricFamilies takes a snapshot of all currently registered
// prometheus MetricFamilies
func SnapshotMetricFamilies() MetricFamilies {
	handler := prometheus.UninstrumentedHandler()
	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	req.Header.Set("Accept", prometheus.DelimitedTelemetryContentType)
	handler.ServeHTTP(recorder, req)

	// We currently don't just do an io.Copy between recorder.Body and a
	// byte buffer since in the future we want whitelist the metrics we
	// send.
	metricFamilies := map[string]*dto.MetricFamily{}
	for {
		mf := &dto.MetricFamily{}
		if _, err = pbutil.ReadDelimited(recorder.Body, mf); err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		metricFamilies[mf.GetName()] = mf
	}
	return metricFamilies
}

// Marshal writes all the metric families to the writer
func (mfs MetricFamilies) Marshal(w io.Writer) error {
	e := expfmt.NewEncoder(w, expfmt.FmtProtoDelim)
	for _, mf := range mfs {
		err := e.Encode(mf)
		if err != nil {
			return err
		}
	}
	return nil
}

// PushToGateway pushes these metrics to the Prometheus PushGateway
func (mfs MetricFamilies) PushToGateway(pushURL, job, instance string) error {
	if !strings.Contains(pushURL, "://") {
		pushURL = "http://" + pushURL
	}
	pushURL = fmt.Sprintf("%s/metrics/jobs/%s/instances/%s", pushURL, url.QueryEscape(job), url.QueryEscape(instance))
	buf := &bytes.Buffer{}
	err := mfs.Marshal(buf)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", pushURL, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", prometheus.DelimitedTelemetryContentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 202 {
		return fmt.Errorf("unexpected status code %d while pushing to %s", resp.StatusCode, pushURL)
	}
	return nil
}

// AnnotateWithClientID updates the labels for each metric to include the
// client_id
func (mfs MetricFamilies) AnnotateWithClientID(clientID string) error {
	for _, mf := range mfs {
		if err := annotateClientID(mf, clientID); err != nil {
			return err
		}
	}
	return nil
}

func annotateClientID(mf *dto.MetricFamily, clientID string) error {
	name := "client_id"
	installationLabel := &dto.LabelPair{
		Name:  &name,
		Value: &clientID,
	}
	for _, metric := range mf.Metric {
		// Sanity check that the customer didn't already specify client_id
		for _, labelPair := range metric.Label {
			if *(labelPair.Name) == name {
				return fmt.Errorf("A metric from client %s already contains the label client_id", clientID)
			}
		}
		metric.Label = append(metric.Label, installationLabel)
		// The labels must be sorted
		sort.Sort(prometheus.LabelPairSorter(metric.Label))
	}
	return nil
}
