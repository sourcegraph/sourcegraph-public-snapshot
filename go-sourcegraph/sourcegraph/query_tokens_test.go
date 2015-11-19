package sourcegraph

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestTokens_JSON(t *testing.T) {
	tokens := Tokens{
		AnyToken("a"),
		Term("b"),
		Term(""),
		RepoToken{URI: "r"},
		RevToken{Rev: "v"},
		FileToken{Path: "p"},
		UserToken{Login: "u"},
	}

	b, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	wantJSON := `[
  {
    "String": "a",
    "Type": "AnyToken"
  },
  {
    "String": "b",
    "Type": "Term"
  },
  {
    "String": "",
    "Type": "Term"
  },
  {
    "uri": "r",
    "Type": "RepoToken"
  },
  {
    "rev": "v",
    "Type": "RevToken"
  },
  {
    "path": "p",
    "Type": "FileToken"
  },
  {
    "login": "u",
    "Type": "UserToken"
  }
]`
	if string(b) != wantJSON {
		t.Errorf("got JSON\n%s\n\nwant JSON\n%s", b, wantJSON)
	}

	var tokens2 Tokens
	if err := json.Unmarshal(b, &tokens2); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(tokens2, tokens) {
		t.Errorf("got tokens\n%+v\n\nwant tokens\n%+v", tokens2, tokens)
	}
}

func TestTokens_nil(t *testing.T) {
	tokens := Tokens(nil)

	b, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	wantJSON := `[]`
	if string(b) != wantJSON {
		t.Errorf("got JSON\n%s\n\nwant JSON\n%s", b, wantJSON)
	}
}

func TestPBToken_Token(t *testing.T) {
	tests := []struct {
		pb   PBToken
		want Token
	}{
		{PBTokenWrap(Term("t")), Term("t")},
		{PBTokenWrap(AnyToken("t")), AnyToken("t")},
		{PBTokenWrap(RepoToken{URI: "r"}), RepoToken{URI: "r"}},
		{PBTokenWrap(RevToken{Rev: "v"}), RevToken{Rev: "v"}},
		{PBTokenWrap(FileToken{Path: "p"}), FileToken{Path: "p"}},
		{PBTokenWrap(UserToken{Login: "u"}), UserToken{Login: "u"}},
		{PBTokenWrap(&RepoToken{URI: "r"}), RepoToken{URI: "r"}},
		{PBTokenWrap(&RevToken{Rev: "v"}), RevToken{Rev: "v"}},
		{PBTokenWrap(&FileToken{Path: "p"}), FileToken{Path: "p"}},
		{PBTokenWrap(&UserToken{Login: "u"}), UserToken{Login: "u"}},

		{PBTokenWrap(nil), Term("")},
		{PBTokenWrap(Term("")), Term("")},
	}
	for _, test := range tests {
		tok := test.pb.GetQueryToken()
		if !reflect.DeepEqual(tok, test.want) {
			t.Errorf("got %#v, want %#v", tok, test.want)
		}
	}
}

func TestPBTokens_JSON(t *testing.T) {
	pbtoks := []PBToken{
		PBTokenWrap(Term("t")),
		PBTokenWrap(AnyToken("t")),
		PBTokenWrap(RepoToken{URI: "r"}),
		PBTokenWrap(RevToken{Rev: "v"}),
		PBTokenWrap(FileToken{Path: "p"}),
		PBTokenWrap(UserToken{Login: "u"}),
		PBTokenWrap(&RepoToken{URI: "r"}),
		PBTokenWrap(&RevToken{Rev: "v"}),
		PBTokenWrap(&FileToken{Path: "p"}),
		PBTokenWrap(&UserToken{Login: "u"}),
	}

	b, err := json.Marshal(pbtoks)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(b))

	var pbtoks2 []PBToken
	if err := json.Unmarshal(b, &pbtoks2); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(pbtoks2, pbtoks) {
		t.Errorf("got %+v, want %+v", pbtoks2, pbtoks)
	}
}
