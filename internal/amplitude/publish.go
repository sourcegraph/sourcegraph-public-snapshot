package amplitude

import (
	"bytes"
	"io"
	"net/http"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
)

// var AMPLITUDE_API_TOKEN = env.Get("AMPLITUDE_API_TOKEN", "", "The API token for the Amplitude project to send data to.")
var AMPLITUDE_API_URL = "https://api2.amplitude.com/2/httpapi"

func Publish(amplitudeAPIToken string, jsonReq []byte) error {

	data := bytes.NewBuffer(jsonReq)
	req, err := http.NewRequest("POST", AMPLITUDE_API_URL, data)
	if err != nil {
		log15.Warn("Could not log Amplitude event", "err", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log15.Warn("Could not log Amplitude event", "err", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			log15.Warn("Could not log Amplitude event", "err", string(buf))
			return err
		}
		log15.Warn("Could not log Amplitude event", "err", string(buf))
		return errors.Errorf("Code %v: %s", resp.StatusCode, string(buf))
	}

	return nil
}
