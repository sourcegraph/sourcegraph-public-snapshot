/*-
 * Copyright 2014 Square Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package jose

import (
	"bytes"
	"testing"
)

func TestBase64URLEncode(t *testing.T) {
	// Test arrays with various sizes
	if base64URLEncode([]byte{}) != "" {
		t.Error("failed to encode empty array")
	}

	if base64URLEncode([]byte{0}) != "AA" {
		t.Error("failed to encode [0x00]")
	}

	if base64URLEncode([]byte{0, 1}) != "AAE" {
		t.Error("failed to encode [0x00, 0x01]")
	}

	if base64URLEncode([]byte{0, 1, 2}) != "AAEC" {
		t.Error("failed to encode [0x00, 0x01, 0x02]")
	}

	if base64URLEncode([]byte{0, 1, 2, 3}) != "AAECAw" {
		t.Error("failed to encode [0x00, 0x01, 0x02, 0x03]")
	}
}

func TestBase64URLDecode(t *testing.T) {
	// Test arrays with various sizes
	val, err := base64URLDecode("")
	if err != nil || !bytes.Equal(val, []byte{}) {
		t.Error("failed to decode empty array")
	}

	val, err = base64URLDecode("AA")
	if err != nil || !bytes.Equal(val, []byte{0}) {
		t.Error("failed to decode [0x00]")
	}

	val, err = base64URLDecode("AAE")
	if err != nil || !bytes.Equal(val, []byte{0, 1}) {
		t.Error("failed to decode [0x00, 0x01]")
	}

	val, err = base64URLDecode("AAEC")
	if err != nil || !bytes.Equal(val, []byte{0, 1, 2}) {
		t.Error("failed to decode [0x00, 0x01, 0x02]")
	}

	val, err = base64URLDecode("AAECAw")
	if err != nil || !bytes.Equal(val, []byte{0, 1, 2, 3}) {
		t.Error("failed to decode [0x00, 0x01, 0x02, 0x03]")
	}
}

func TestDeflateRoundtrip(t *testing.T) {
	original := []byte("Lorem ipsum dolor sit amet")

	compressed, err := deflate(original)
	if err != nil {
		panic(err)
	}

	output, err := inflate(compressed)
	if err != nil {
		panic(err)
	}

	if bytes.Compare(output, original) != 0 {
		t.Error("Input and output do not match")
	}
}

func TestInvalidCompression(t *testing.T) {
	_, err := compress("XYZ", []byte{})
	if err == nil {
		t.Error("should not accept invalid algorithm")
	}

	_, err = decompress("XYZ", []byte{})
	if err == nil {
		t.Error("should not accept invalid algorithm")
	}

	_, err = decompress(DEFLATE, []byte{1, 2, 3, 4})
	if err == nil {
		t.Error("should not accept invalid data")
	}
}
