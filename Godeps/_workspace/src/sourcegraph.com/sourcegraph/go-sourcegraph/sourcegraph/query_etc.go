package sourcegraph

import (
	"encoding/json"
	"fmt"
)

func (e TokenError) Error() string { return fmt.Sprintf("%s (%v)", e.Message, e.Token) }

type jsonTokenError struct {
	Index   int        `json:",omitempty"`
	Token   *jsonToken `json:",omitempty"`
	Message string
}

func (e TokenError) MarshalJSON() ([]byte, error) {
	var tok *jsonToken
	if e.Token != nil {
		tok = &jsonToken{e.Token.GetQueryToken()}
	}
	return json.Marshal(jsonTokenError{int(e.Index), tok, e.Message})
}

func (e *TokenError) UnmarshalJSON(b []byte) error {
	var jv jsonTokenError
	if err := json.Unmarshal(b, &jv); err != nil {
		return err
	}

	var tok *PBToken
	if jv.Token != nil {
		tmp := PBTokenWrap(jv.Token.Token)
		tok = &tmp
	}

	*e = TokenError{int32(jv.Index), tok, jv.Message}
	return nil
}
