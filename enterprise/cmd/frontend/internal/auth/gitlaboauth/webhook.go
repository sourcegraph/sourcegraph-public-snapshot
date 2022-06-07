package gitlaboauth

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type userAccountRequest struct {
	Email string `json:"email"`
}

func UserStatusFromWebhook(url string, email string) (int, error) {
	user := userAccountRequest{
		Email: email,
	}

	body, _ := json.Marshal(user)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}

	return req.Response.StatusCode, err
}
