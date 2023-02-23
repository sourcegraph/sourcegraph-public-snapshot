package srcprometheus

import (
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

type AlertsStatus struct {
	Warning          int `json:"warning"`
	Silenced         int `json:"silenced"`
	Critical         int `json:"critical"`
	ServicesCritical int `json:"services_critical"`
}

type ConfigStatus struct {
	Problems conf.Problems `json:"problems"`
}
