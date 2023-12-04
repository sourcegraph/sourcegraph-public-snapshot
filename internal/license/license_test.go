package license

import (
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/license/licensetest"
)

func TestParseTagsInput(t *testing.T) {
	tests := map[string][]string{}
	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			got := ParseTagsInput(input)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}

var (
	timeFixture   = time.Date(2018, time.September, 22, 21, 33, 44, 0, time.UTC)
	infoV1Fixture = Info{Tags: []string{"a"}, UserCount: 123, ExpiresAt: timeFixture}

	sfSubID       = "AE0002412312"
	sfOpID        = "EA890000813"
	infoV2Fixture = Info{Tags: []string{"a"}, UserCount: 123, ExpiresAt: timeFixture, SalesforceSubscriptionID: &sfSubID, SalesforceOpportunityID: &sfOpID}
)

func TestInfo_EncodeDecode(t *testing.T) {
	t.Run("v1 ok", func(t *testing.T) {
		want := infoV1Fixture
		data, err := want.encode()
		if err != nil {
			t.Fatal(err)
		}

		var got Info
		if err := got.decode(data); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("v2 ok", func(t *testing.T) {
		want := infoV2Fixture
		data, err := want.encode()
		if err != nil {
			t.Fatal(err)
		}

		var got Info
		if err := got.decode(data); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		var got Info
		if err := got.decode([]byte("invalid")); err == nil {
			t.Fatal("want error")
		}
	})
}

func TestGenerateParseSignedKey(t *testing.T) {
	t.Run("v1 ok", func(t *testing.T) {
		want := infoV1Fixture
		text, _, err := GenerateSignedKey(want, licensetest.PrivateKey)
		if err != nil {
			t.Fatal(err)
		}

		got, _, err := ParseSignedKey(text, licensetest.PublicKey)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(got, &want) {
			t.Errorf("got %+v, want %+v", got, &want)
		}
	})

	t.Run("v2 ok", func(t *testing.T) {
		want := infoV2Fixture
		text, _, err := GenerateSignedKey(want, licensetest.PrivateKey)
		if err != nil {
			t.Fatal(err)
		}

		got, _, err := ParseSignedKey(text, licensetest.PublicKey)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(got, &want) {
			t.Errorf("got %+v, want %+v", got, &want)
		}
	})

	t.Run("ignores whitespace", func(t *testing.T) {
		want := infoV1Fixture
		text, _, err := GenerateSignedKey(want, licensetest.PrivateKey)
		if err != nil {
			t.Fatal(err)
		}

		// Add some whitespace.
		text = text[:20] + " \n \t" + text[20:40] + " " + text[40:]

		got, _, err := ParseSignedKey(text, licensetest.PublicKey)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(got, &want) {
			t.Errorf("got %+v, want %+v", got, &want)
		}
	})
}

func TestParseSignedKey(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		if _, _, err := ParseSignedKey("invalid", licensetest.PublicKey); err == nil {
			t.Fatal("want error")
		}
	})
}
