package amplitude

import (
	"errors"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var AMPLITUDE_API_TOKEN = env.Get("AMPLITUDE_API_TOKEN", "", "The API token for the Amplitude project to send data to.")
var AMPLITUDE_API_URL = "https://api2.amplitude.com/2/httpapi"

func Publish(data string) error {
	resp, err := http.Post(AMPLITUDE_API_URL, "application/json", nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("Amplitude Error")
	}
	return nil
}
