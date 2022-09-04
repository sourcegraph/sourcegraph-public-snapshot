package webapp

import (
	"fmt"
	"net/http"
)

func (h *Handler) serveInstances(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello world")
}
