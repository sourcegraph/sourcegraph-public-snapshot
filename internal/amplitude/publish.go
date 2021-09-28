package amplitude

import (
	"bytes"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
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
	resp, err := httpcli.ExternalDoer.Do(req)
	if err != nil {
		return errors.WithMessage(err, "amplitude: cannot create new request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusBadRequest:
			log15.Warn("amplitude: Could not log event: JSON formatting incorrect.", "err", err)
		case http.StatusRequestEntityTooLarge:
			// We should never hit this, because we send a single event at a time.
			// Notify the user, but a TODO is to properly handle retries for this case.
			log15.Warn("amplitude: Could not log event: Payload too large.", "err", err)
		case http.StatusTooManyRequests:
			// Amplitude may throttle us if we exceed 1000 events/sec. Give a 30 second break before retrying.
			log15.Warn("amplitude: Could not log event: Too many requests. Maximum 10 events/second/user. Retrying in 30s.", "err", err)
		case http.StatusInternalServerError:
			log15.Warn("amplitude: Could not log event: Internal server error.", "err", err)
		}
	}

	return nil
}
