package metricutil

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/gogo/protobuf/proto"
	dto "github.com/prometheus/client_model/go"
)

func TestMarshalUnmarshal(t *testing.T) {
	wantedFamilies := testMetricsFamilies()
	writeBuf := &bytes.Buffer{}
	wantedFamilies.Marshal(writeBuf)
	readBuf := bytes.NewBuffer(writeBuf.Bytes())
	gotFamilies := UnmarshalMetricFamilies(readBuf)
	if !reflect.DeepEqual(wantedFamilies, gotFamilies) {
		t.Fatal("x != Unmarshal(Marshal(x))")
	}
}

func TestAnnotateWithClientID(t *testing.T) {
	mfs := testMetricsFamilies()
	expectedClientID := "test"
	error := mfs.AnnotateWithClientID(expectedClientID)
	if error != nil {
		t.Fatal("Error annotating client_id")
	}
	if reflect.DeepEqual(mfs, testMetricsFamilies) {
		t.Fatal("AnnotateWithClientID did not mutate")
	}
	for _, mf := range mfs {
		for _, m := range mf.Metric {
			gotClientID := ""
			for _, label := range m.Label {
				if *(label.Name) == "client_id" {
					gotClientID = *(label.Value)
				}
			}
			if expectedClientID != gotClientID {
				t.Fatal("m.Name client id mismatch", expectedClientID, gotClientID)
			}
		}
	}
}

func TestAnnotateWithClientIDFail(t *testing.T) {
	mfs := testMetricsFamilies()
	labelName := "client_id"
	// Modify input such that we have a client_id already set
	for _, mf := range mfs {
		for _, m := range mf.Metric {
			for _, label := range m.Label {
				if *(label.Name) == "label" {
					label.Name = &labelName
				}
			}
		}
	}
	error := mfs.AnnotateWithClientID("test")
	if error == nil {
		t.Fatal("Expected error while annotating")
	}
}

func testMetricsFamilies() MetricFamilies {
	return MetricFamilies{
		"foo": &dto.MetricFamily{
			Name: proto.String("foo"),
			Help: proto.String("foodoc"),
			Type: dto.MetricType_COUNTER.Enum(),
			Metric: []*dto.Metric{
				{
					Label: []*dto.LabelPair{
						{
							Name:  proto.String("label"),
							Value: proto.String("hello"),
						},
						{
							Name:  proto.String("type"),
							Value: proto.String("foo"),
						},
					},
					Counter: &dto.Counter{
						Value: proto.Float64(1),
					},
				},
			},
		},
		"bar": &dto.MetricFamily{
			Name: proto.String("bar"),
			Help: proto.String("bardoc"),
			Type: dto.MetricType_COUNTER.Enum(),
			Metric: []*dto.Metric{
				{
					Label: []*dto.LabelPair{
						{
							Name:  proto.String("label"),
							Value: proto.String("hello"),
						},
						{
							Name:  proto.String("type"),
							Value: proto.String("bar"),
						},
					},
					Counter: &dto.Counter{
						Value: proto.Float64(2),
					},
				},
			},
		},
	}
}
