package license

import (
	"reflect"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
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
	privateKey ssh.Signer
	publicKey  ssh.PublicKey
)

func init() {
	// This private key is used for testing only. It is not the private key used to generate valid
	// Sourcegraph license keys.
	const (
		testPrivateKeyData = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAr+TgUgYjW5aK4+St92YVYEKlxkH8uuJdqcGdDL4hxPaW8y7T
7hSPRwCghh/s8+lo/KF7OcwQOILSdc+QiS9zLAMOyuwaIpa67lxD+NU1GRLiPAc0
gtci3cRCLxPknDq+r9IsqUdKc67UEEKYJmTW19IIkRwD7cqrMoW3hTy7PrKyfx1z
AmQY9XoHdOT+7F8UO51MLpJVwwv39iX93m6y9tJ8Oub8XT0oOddioYzWJXNUeInZ
9aWQ3E8EpaFhuhc9N9pqYbiSeqWdwl2SIqP0qD6C1Th8c2Fr3A99MYBooBQ3er+E
kEZiQFpiy8YBXXfzyp8OCoyADTC/htTbyaWCrwIDAQABAoIBAQCcVMgrHtl+Jab8
NduI72pAS/wa4btGPrsQjcyP7s0EykYPjac/C3bsnFoP24A2qVVuc/eZvw0LrPpx
SzIjO9EZZP5YzM/NvjoWcFrZZmXzCx2YTK8XOy75+9z3Cq89S9j/W8XmDj7V4qUI
bqfcN/PbjgmfL08uodrW5JHgEcI6Tg5JM0jKM8AmQ5PdbFYq5QHycYtg1jrKsfyN
bT1omoFl5DDD8JSY9odsJF3zcoaD0tOAm+IELKCj/jW04OpDW5daUmVa0fC3EAc8
zrKpTrdHhT32ox05f+tup8CCsAz2jgrl4Q493x+idpjySqOfgdZJpIYl9fINyXtB
uuG/VL+BAoGBAOdi78/mBlxvmeAJ/0P5Hq46aVaPr5N+cT61iyMJErWVhiXYzmWQ
KU4X48OT7jHli3AnkxuKznq7+K484ld2TQAa15QAes5ncTTosptkI7G1W+PuvtnW
dZUT4KYQQO29q5yqDA6V0cTBL9ErM5JxQHjXfn8HGJZrZwjgmL16csZpAoGBAMKa
yFUaD5C9aXrHDQX8KxH5hdJPPcWsqDF3jVdGMCvuH+at5SXKQEHpyNKHPG5bSiHC
x4wsZuekXj22mLguvRO0OgFAi6o6XqTq3sHCC88as59yID3S+Yebnwoi5eTT39Nw
1mRMLdBdxaUIGjr7YcCWHdostm9vP1qxG4KBcs1XAoGAJsQosYg0YKBCA4spPVYr
kb0vkNUZ8XgpuPvph8EpZUrb4tNkIRf6E59lBYtvSOTQa8X1m5Ox7FY539dLhKPJ
Ws8kdyNtb89c5CRoaChq4ockEhgZ2Z1YrdVauffXKP2yFtlWk8r+DhzfygeW6F4Z
EyXnh5jEwc7UwGQAsx4kxKECgYBVe5BMcaostkdDO3SpEGALAUSbYVuNHY60TAO2
NoqqgWtZ9UEXHISlb4BbmjQdddRWiC0HkemSD02mZjMnlMBRi3V/X076c7FnWBSJ
oCd3zg7hF0y6W5Zozx9JYJMDtV44jvReEmh9gvNyzuBW0F3cLxTl5YYt5Pw7Zljj
NuGq+QKBgB/JFOlfywL6IiMAqN90THg7tbbEZZXO+hylXE5kItQf2aPH6H1iKyhE
yww3YER5jCVAhbgotx4PrPMyjRqs+SrUtvDaqAlhY0YDHUFiAqk/Rukx4EfQlOT9
mSXt7lUbEmiQep700eM7YlgrOxUVqHsjf1QMrNfq05Ajr8uDfHim
-----END RSA PRIVATE KEY-----`
		testPublicKeyData = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCv5OBSBiNblorj5K33ZhVgQqXGQfy64l2pwZ0MviHE9pbzLtPuFI9HAKCGH+zz6Wj8oXs5zBA4gtJ1z5CJL3MsAw7K7BoilrruXEP41TUZEuI8BzSC1yLdxEIvE+ScOr6v0iypR0pzrtQQQpgmZNbX0giRHAPtyqsyhbeFPLs+srJ/HXMCZBj1egd05P7sXxQ7nUwuklXDC/f2Jf3ebrL20nw65vxdPSg512KhjNYlc1R4idn1pZDcTwSloWG6Fz032mphuJJ6pZ3CXZIio/SoPoLVOHxzYWvcD30xgGigFDd6v4SQRmJAWmLLxgFdd/PKnw4KjIANML+G1NvJpYKv`
	)

	var err error
	privateKey, err = ssh.ParsePrivateKey([]byte(testPrivateKeyData))
	if err != nil {
		panic(err)
	}
	publicKey, _, _, _, err = ssh.ParseAuthorizedKey([]byte(testPublicKeyData))
	if err != nil {
		panic(err)
	}
}

var (
	timeFixture = time.Date(2018, time.September, 22, 21, 33, 44, 0, time.UTC)
	infoFixture = Info{Tags: []string{"a"}, UserCount: 123, ExpiresAt: timeFixture}
)

func TestInfo_EncodeDecode(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		want := infoFixture
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
	t.Run("ok", func(t *testing.T) {
		want := infoFixture
		text, err := GenerateSignedKey(want, privateKey)
		if err != nil {
			t.Fatal(err)
		}

		got, _, err := ParseSignedKey(text, publicKey)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(got, &want) {
			t.Errorf("got %+v, want %+v", got, &want)
		}
	})

	t.Run("ignores whitespace", func(t *testing.T) {
		want := infoFixture
		text, err := GenerateSignedKey(want, privateKey)
		if err != nil {
			t.Fatal(err)
		}

		// Add some whitespace.
		text = text[:20] + " \n \t" + text[20:40] + " " + text[40:]

		got, _, err := ParseSignedKey(text, publicKey)
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
		if _, _, err := ParseSignedKey("invalid", publicKey); err == nil {
			t.Fatal("want error")
		}
	})
}
