package sourcegraph

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
)

// A Token is the smallest indivisible component of a query, either a
// term or a "field:val" specifier (e.g., "repo:example.com/myrepo").
type Token interface {
	// Token returns the string representation of the token.
	Token() string
}

// A Term is a query term token. It is either a word or an arbitrary
// string (if quoted in the raw query).
type Term string

func (t Term) Token() string {
	if strings.Contains(string(t), " ") {
		return `"` + string(t) + `"`
	}
	return string(t)
}

func (t Term) UnquotedToken() string { return string(t) }

// An AnyToken is a token that has not yet been resolved into another
// token type. It resolves to Term if it can't be resolved to another
// token type.
type AnyToken string

func (u AnyToken) Token() string { return string(u) }

func (t RepoToken) Token() string { return t.URI }

func (t RepoToken) Spec() RepoSpec {
	return RepoSpec{URI: t.URI}
}

func (t RevToken) Token() string { return ":" + t.Rev }

func (t UnitToken) Token() string {
	s := "~" + t.Name
	if t.UnitType != "" {
		s += "@" + t.UnitType
	}
	return s
}

func (t FileToken) Token() string { return "/" + filepath.Clean(t.Path) }

func (t UserToken) Token() string { return "@" + t.Login }

// Tokens wraps a list of tokens and adds some helper methods. It also
// serializes to JSON with "Type" fields added to each token and
// deserializes that same JSON back into a typed list of tokens.
type Tokens []Token

func (d Tokens) MarshalJSON() ([]byte, error) {
	jtoks := make([]jsonToken, len(d))
	for i, t := range d {
		jtoks[i] = jsonToken{t}
	}
	return json.Marshal(jtoks)
}

func (d *Tokens) UnmarshalJSON(b []byte) error {
	var jtoks []jsonToken
	if err := json.Unmarshal(b, &jtoks); err != nil {
		return err
	}
	if jtoks == nil {
		*d = nil
	} else {
		*d = make(Tokens, len(jtoks))
		for i, jtok := range jtoks {
			(*d)[i] = jtok.Token
		}
	}
	return nil
}

func (d Tokens) RawQueryString() string { return Join(d).Text }

type jsonToken struct {
	Token `json:",omitempty"`
}

func (t jsonToken) MarshalJSON() ([]byte, error) {
	if t.Token == nil {
		return []byte("null"), nil
	}
	tokType := TokenType(t.Token)
	b, err := json.Marshal(t.Token)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return b, nil
	}
	switch b[0] {
	case '"':
		b = []byte(fmt.Sprintf(`{"String":%s,"Type":%q}`, b, tokType))
	case '{':
		b[len(b)-1] = ','
		b = append(b, []byte(fmt.Sprintf(`"Type":%q}`, tokType))...)
	}
	return b, nil
}

func (t *jsonToken) UnmarshalJSON(b []byte) error {
	var typ struct{ Type string }
	if err := json.Unmarshal(b, &typ); err != nil {
		return err
	}
	if typ.Type == "" {
		return nil
	}

	*t = jsonToken{}
	switch typ.Type {
	case "":
		return nil
	case "Term", "AnyToken":
		var tmp struct{ String string }
		if err := json.Unmarshal(b, &tmp); err != nil {
			return err
		}
		switch typ.Type {
		case "Term":
			t.Token = Term(tmp.String)
		case "AnyToken":
			t.Token = AnyToken(tmp.String)
		}
		return nil
	case "RepoToken":
		t.Token = &RepoToken{}
	case "RevToken":
		t.Token = &RevToken{}
	case "UnitToken":
		t.Token = &UnitToken{}
	case "FileToken":
		t.Token = &FileToken{}
	case "UserToken":
		t.Token = &UserToken{}
	default:
		return fmt.Errorf("unmarshal Tokens: unrecognized Type %q", typ.Type)
	}
	if err := json.Unmarshal(b, t.Token); err != nil {
		return err
	}
	t.Token = reflect.ValueOf(t.Token).Elem().Interface().(Token)
	return nil
}

func TokenType(tok Token) string {
	return strings.Replace(strings.Replace(reflect.ValueOf(tok).Type().String(), "*", "", -1), "sourcegraph.", "", -1)
}

func (t PBToken) MarshalJSON() ([]byte, error) {
	jt := jsonToken{t.Token()}
	return jt.MarshalJSON()
}

func (t *PBToken) UnmarshalJSON(b []byte) error {
	var jt jsonToken
	if err := json.Unmarshal(b, &jt); err != nil {
		return err
	}
	*t = PBTokenWrap(jt.Token)
	return nil
}

// Token returns the Token that the PBToken wraps.
func (t *PBToken) Token() Token {
	switch {
	case t.Term != "":
		return Term(t.Term)
	case t.AnyToken != "":
		return AnyToken(t.AnyToken)
	case t.RepoToken != nil:
		return *t.RepoToken
	case t.RevToken != nil:
		return *t.RevToken
	case t.FileToken != nil:
		return *t.FileToken
	case t.UserToken != nil:
		return *t.UserToken
	case t.UnitToken != nil:
		return *t.UnitToken
	default:
		// empty
		return Term("")
	}
}

func PBTokenWrap(t Token) PBToken {
	var pb PBToken
	switch t := t.(type) {
	case Term:
		pb.Term = string(t)
	case AnyToken:
		pb.AnyToken = string(t)
	case RepoToken:
		pb.RepoToken = &t
	case RevToken:
		pb.RevToken = &t
	case FileToken:
		pb.FileToken = &t
	case UserToken:
		pb.UserToken = &t
	case UnitToken:
		pb.UnitToken = &t
	case *RepoToken:
		pb.RepoToken = t
	case *RevToken:
		pb.RevToken = t
	case *FileToken:
		pb.FileToken = t
	case *UserToken:
		pb.UserToken = t
	case *UnitToken:
		pb.UnitToken = t
	default:
		// empty
	}
	return pb
}

func PBTokensWrap(toks []Token) []PBToken {
	pbtoks := make([]PBToken, len(toks))
	for i, tok := range toks {
		pbtoks[i] = PBTokenWrap(tok)
	}
	return pbtoks
}

// PBTokens converts []PBToken to Tokens.
func PBTokens(pbtoks []PBToken) Tokens {
	toks := make(Tokens, len(pbtoks))
	for i, pbtok := range pbtoks {
		toks[i] = pbtok.Token()
	}
	return toks
}
