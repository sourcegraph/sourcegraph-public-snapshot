package amplitude

import (
	"bytes"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
)

const apiUrl = "https://api2.amplitude.com/2/httpapi"

func Publish(body []byte) error {
	data := bytes.NewBuffer(body)
	req, err := http.NewRequest("POST", apiUrl, data)
	if err != nil {
		log15.Warn("Could not log Amplitude event", "err", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	client := &http.Client{}
	attempts := 0
	for {
		resp, err := client.Do(req)
		if err != nil {
			if resp.StatusCode == http.StatusBadRequest {
				log15.Warn("Could not log Amplitude event: JSON formatting incorrect.", "err", err)
			}
			if resp.StatusCode == http.StatusRequestEntityTooLarge {
				// We should never hit this, because we send a single event at a time.
				// Notify the user, but a TODO is to properly handle retries for this case.
				log15.Warn("Could not log Amplitude event: Payload too large.", "err", err)
			}
			if resp.StatusCode == http.StatusTooManyRequests {
				// Amplitude may throttle us if we exceed 1000 events/sec. Give a 30 second break before retrying.
				log15.Warn("Could not log Amplitude event: Too many requests. Maximum 10 events/second/user. Retrying in 30s.", "err", err)
				time.Sleep(30 * time.Second)
			}
			if resp.StatusCode == http.StatusInternalServerError {
				log15.Warn("Could not log Amplitude event: Internal server error.", "err", err)
				time.Sleep(5 * time.Second)
			}
			if attempts > 5 {
				log15.Warn("Could not log Amplitude event. Not retrying", "err", err)
				return errors.Errorf("Could not log Amplitude event. Not retrying. Code: %v, Message: %v", resp.StatusCode, err.Error())
			}
			attempts++
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		}

	}
}
