package amplitude

import (
	"bytes"
	"io"
	"net/http"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

const apiURL = "https://api2.amplitude.com/2/httpapi"

// Publish publishes an event to the Amplitude project.
func Publish(body []byte) error {
	data := bytes.NewBuffer(body)
	req, err := http.NewRequest("POST", apiURL, data)
	if err != nil {
		return errors.WithMessage(err, "amplitude: Cannot create new request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	resp, err := httpcli.ExternalDoer.Do(req)
	if err != nil {
		return errors.WithMessage(err, "amplitude: Could not log Amplitude event")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		switch resp.StatusCode {
		case http.StatusBadRequest:
			return errors.WithMessage(err, "amplitude: Could not log event: JSON formatting incorrect.")
		case http.StatusRequestEntityTooLarge:
			// We should never hit this, because we send a single event at a time.
			// Notify the user, but a TODO is to properly handle retries for this case.
			return errors.WithMessage(err, "amplitude: Could not log event: Payload too large.")
		case http.StatusTooManyRequests:
			// Amplitude may throttle us if we exceed 1000 events/sec.
			return errors.WithMessage(err, "amplitude: Could not log event: Too many requests. Maximum 10 events/second/user.")
		case http.StatusInternalServerError:
			return errors.WithMessage(err, "amplitude: Could not log event: Internal server error.")
		default:
			return errors.Errorf("amplitude: Failed with %d %s", resp.StatusCode, string(body))
		}
	}

	return nil
}
