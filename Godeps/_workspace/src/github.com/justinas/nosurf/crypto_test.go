package nosurf

import (
	"bytes"
	"testing"
)

func TestOtpPanicsOnLengthMismatch(t *testing.T) {
	data := make([]byte, 1)
	key := make([]byte, 2)

	defer func() {
		if r := recover(); r == nil {
			t.Error("One time pad should've panicked on receiving slices" +
				"of different length, but it didn't")
		}
	}()
	oneTimePad(data, key)
}
func TestOtpMasksCorrectly(t *testing.T) {
	data := []byte("Inventors of the shish-kebab")
	key := []byte("They stop Cthulhu eating ye.")
	// precalculated
	expected := []byte("\x1d\x06\x13\x1cN\x07\x1b\x1d\x03\x00,\x12H\x01\x04" +
		"\rUS\r\x08\x07\x01C\x0cE\x1b\x04L")

	oneTimePad(data, key)

	if !bytes.Equal(data, expected) {
		t.Errorf("oneTimePad masked the data incorrectly: expected %#v, got %#v",
			expected, data)
	}
}

func TestOtpUnmasksCorrectly(t *testing.T) {
	orig := []byte("a very secret message")
	data := make([]byte, len(orig))
	copy(data, orig)
	if !bytes.Equal(orig, data) {
		t.Fatal("copy failed")
	}

	key := []byte("even more secret key!")

	oneTimePad(data, key)
	oneTimePad(data, key)

	if !bytes.Equal(orig, data) {
		t.Errorf("2x oneTimePad didn't return the original data:"+
			" expected %#v, got %#v", orig, data)
	}
}

func TestMasksTokenCorrectly(t *testing.T) {
	// needs to be of tokenLength
	token := []byte("12345678901234567890123456789012")
	fullToken := maskToken(token)

	if len(fullToken) != 2*tokenLength {
		t.Errorf("len(fullToken) is not %d, but %d", 2*tokenLength, len(fullToken))
	}

	key := fullToken[:tokenLength]
	encToken := fullToken[tokenLength:]

	// perform unmasking
	oneTimePad(encToken, key)

	if !bytes.Equal(encToken, token) {
		t.Errorf("Unmasked token is invalid: expected %v, got %v", token, encToken)
	}
}

func TestUnmasksTokenCorrectly(t *testing.T) {
	token := []byte("12345678901234567890123456789012")
	fullToken := maskToken(token)

	decToken := unmaskToken(fullToken)

	if !bytes.Equal(decToken, token) {
		t.Errorf("Unmasked token is invalid: expected %v, got %v", token, decToken)
	}
}
