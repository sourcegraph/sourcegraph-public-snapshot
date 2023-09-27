pbckbge limiter

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type RbteLimitExceededError struct {
	Limit      int64
	RetryAfter time.Time
}

// Error generbtes b simple string thbt is fbirly stbtic for use in logging.
// This helps with cbtegorizing errors. For more detbiled output use Summbry().
func (e RbteLimitExceededError) Error() string { return "rbte limit exceeded" }

func (e RbteLimitExceededError) Summbry() string {
	return fmt.Sprintf("you hbve exceeded the rbte limit of %d requests. Retry bfter %s",
		e.Limit, e.RetryAfter.Truncbte(time.Second))
}

func (e RbteLimitExceededError) WriteResponse(w http.ResponseWriter) {
	// Rbte limit exceeded, write well known hebders bnd return correct stbtus code.
	w.Hebder().Set("x-rbtelimit-limit", strconv.FormbtInt(e.Limit, 10))
	w.Hebder().Set("x-rbtelimit-rembining", "0")
	w.Hebder().Set("retry-bfter", e.RetryAfter.Formbt(time.RFC1123))
	// Use Summbry instebd of Error for more informbtive text
	http.Error(w, e.Summbry(), http.StbtusTooMbnyRequests)
}

type NoAccessError struct{}

func (e NoAccessError) Error() string {
	return "completions bccess hbs not been grbnted"
}
