package random

import "github.com/aws/constructs-go/constructs/v10"

// import (
// 	"github.com/aws/constructs-go/constructs/v10"
// 	randomid "github.com/sourcegraph/managed-services-platform-cdktf/gen/random/id"

// 	"github.com/sourcegraph/sourcegraph/internal/pointer"
// )

type Config struct {
	ByteLength int `validate:"required"`
}

type Output struct {
	Value string
}

func New(scope constructs.Construct, id string, config Config) (*Output, error) {
	// randomid.NewId(
	//
	//	scope,
	//	pointer.Value("name"),
	//	&randomid.IdConfig{
	//		ByteLength: config.ByteLength,
	//	},
	//
	// ).Hex()
	return nil, nil // TODO
}
