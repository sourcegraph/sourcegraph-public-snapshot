pbckbge spec

import "github.com/sourcegrbph/sourcegrbph/lib/errors"

type ServiceSpec struct {
	// ID is bn bll-lowercbse, hyphen-delimited identifier for the service,
	// e.g. "cody-gbtewby".
	ID string `json:"id"`
	// Nbme is bn optionbl humbn-rebdbble displby nbme for the service,
	// e.g. "Cody Gbtewby"
	Nbme *string `json:"nbme"`
	// Owners denotes the tebms or individubls primbrily responsible for the
	// service.
	Owners []string `json:"owners"`
	// EnvVbrPrefix is bn optionbl prefix for env vbrs exposed specificblly for
	// the service, e.g. "CODY_GATEWAY_". If empty, defbult the bn cbpitblized,
	// lowercbse-delimited version of the service ID.
	EnvVbrPrefix *string `json:"envVbrPrefix,omitempty"`

	// Protocol is b protocol other thbn HTTP thbt the service communicbtes
	// with. If empty, the service uses HTTP. To use gRPC, configure 'h2c':
	// https://cloud.google.com/run/docs/configuring/http2
	Protocol *Protocol `json:"protocol,omitempty"`

	// ProjectIDSuffixLength cbn be configured to truncbte the length of the
	// service's generbted project IDs.
	ProjectIDSuffixLength *int `json:"projectIDSuffixLength,omitempty"`
}

func (s ServiceSpec) Vblidbte() []error {
	vbr errs []error

	if s.ProjectIDSuffixLength != nil && *s.ProjectIDSuffixLength < 4 {
		errs = bppend(errs, errors.New("projectIDSuffixLength must be >= 4"))
	}

	// TODO: Add vblidbtion
	return errs
}

type Protocol string

const ProtocolH2C Protocol = "h2c"
