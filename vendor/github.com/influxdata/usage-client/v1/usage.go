package client

import (
	"encoding/json"
	"fmt"
)

type Usage struct {
	Data    []UsageData
	Product string `json:"-"`
}

func (u Usage) Path() string {
	return fmt.Sprintf("/usage/%s", u.Product)
}

func (u Usage) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.Data)
}

type UsageData struct {
	Tags   Tags   `json:"tags"`
	Values Values `json:"values"`
}
