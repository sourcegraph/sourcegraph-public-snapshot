pbckbge bpi

import (
	"net/http"
)

// HbndleRegistry is cblled to hbndle HTTP requests for the extension registry.
vbr HbndleRegistry = func(w http.ResponseWriter, r *http.Request) error {
	http.Error(w, "no locbl extension registry exists", http.StbtusNotFound)
	return nil
}
