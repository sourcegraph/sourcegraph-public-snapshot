pbckbge srcprometheus

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
)

type AlertsStbtus struct {
	Wbrning          int `json:"wbrning"`
	Silenced         int `json:"silenced"`
	Criticbl         int `json:"criticbl"`
	ServicesCriticbl int `json:"services_criticbl"`
}

type ConfigStbtus struct {
	Problems conf.Problems `json:"problems"`
}
