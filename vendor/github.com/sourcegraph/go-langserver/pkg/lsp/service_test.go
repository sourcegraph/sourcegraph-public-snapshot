package lsp

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestTextDocumentSyncOptionsOrKind_MarshalUnmarshalJSON(t *testing.T) {
	kindPtr := func(kind TextDocumentSyncKind) *TextDocumentSyncKind {
		return &kind
	}

	tests := []struct {
		data []byte
		want TextDocumentSyncOptionsOrKind
	}{
		{
			data: []byte(`null`),
			want: TextDocumentSyncOptionsOrKind{},
		},
		{
			data: []byte(`2`),
			want: TextDocumentSyncOptionsOrKind{
				Options: &TextDocumentSyncOptions{
					OpenClose: true,
					Change:    TDSKIncremental,
				},
				Kind: kindPtr(2),
			},
		},
		{
			data: []byte(`{"openClose":true,"change":1,"save":{"includeText":true}}`),
			want: TextDocumentSyncOptionsOrKind{
				Options: &TextDocumentSyncOptions{
					OpenClose: true,
					Change:    TDSKFull,
					Save:      &SaveOptions{IncludeText: true},
				},
			},
		},
	}
	for _, test := range tests {
		var got TextDocumentSyncOptionsOrKind
		if err := json.Unmarshal(test.data, &got); err != nil {
			t.Error(err)
			continue
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("got %+v, want %+v", got, test.want)
			continue
		}
		data, err := json.Marshal(got)
		if err != nil {
			t.Error(err)
			continue
		}
		if !bytes.Equal(data, test.data) {
			t.Errorf("got JSON %q, want %q", data, test.data)
		}
	}
}
