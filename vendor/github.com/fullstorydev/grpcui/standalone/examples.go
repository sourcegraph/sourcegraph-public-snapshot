package standalone

import (
	"encoding/json"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/golang/protobuf/proto" //lint:ignore SA1019 temporarily still supporting generated code from old plugin
	protov2 "google.golang.org/protobuf/proto"
)

// Example model of an example gRPC request
type Example struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Service     string         `json:"service"`
	Method      string         `json:"method"`
	Request     ExampleRequest `json:"request"`
}

// ExampleMetadataPair (name, value) pair
type ExampleMetadataPair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ExampleRequest gRPC request
type ExampleRequest struct {
	TimeoutSeconds float64               `json:"timeout_secs"`
	Metadata       []ExampleMetadataPair `json:"metadata"`
	Data           interface{}           `json:"data"`
}

func (r *ExampleRequest) MarshalJSON() ([]byte, error) {
	// we override marshaling so that we can correctly handle instances
	// of proto.Message in the Data field
	marshalData, err := marshalData(r.Data)
	if err != nil {
		return nil, err
	}
	// need new named type so next step uses struct reflection to marshal
	//  instead of re-invoking this method
	type jsReq ExampleRequest
	jsonRequest := jsReq{
		TimeoutSeconds: r.TimeoutSeconds,
		Metadata:       r.Metadata,
		Data:           marshalData,
	}

	return json.Marshal(jsonRequest)
}

func marshalData(data interface{}) (json.RawMessage, error) {
	switch data := data.(type) {
	case protov2.Message:
		return protojson.Marshal(data)
	case proto.Message:
		return protojson.Marshal(proto.MessageV2(data))
	default:
		return json.Marshal(data)
	}
}
