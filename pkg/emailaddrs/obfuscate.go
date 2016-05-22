package emailaddrs

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"strings"
)

var emailObfuscationKey, emailObfuscationIV []byte

func init() {
	// emailObfuscationKeyBase64 is the AES encryption key for obfuscating email
	// addresses, and emailObfuscationIVBase64 is the CTR mode IV. Email
	// obfuscation is not security sensitive, so it's relatively harmless to
	// check these secrets into the git repository for convenience.
	const (
		emailObfuscationKeyBase64 = "7TKBNMca8/J9tAfTzz5/PQ=="
		emailObfuscationIVBase64  = "ZTNJNjBGD2c5x9+DeEYQhQ=="
	)

	var err error
	emailObfuscationKey, err = base64.StdEncoding.DecodeString(emailObfuscationKeyBase64)
	if err != nil {
		panic("failed to decode emailObfuscationKeyBase64: " + err.Error())
	}
	emailObfuscationIV, err = base64.StdEncoding.DecodeString(emailObfuscationIVBase64)
	if err != nil {
		panic("failed to decode emailObfuscationKeyIV: " + err.Error())
	}
}

const ObfuscatedEmailDomainPrefix = "-x-"

// Obfuscate scrambles an email address to protect the privacy of the
// holder. It is a tame anti-spam and privacy technique but should not
// be relied on for secrecy.
func Obfuscate(email string) (obfuscated string, err error) {
	user, domain, err := Split(email)
	if err != nil {
		return "", err
	}
	return user + "@" + ObfuscatedEmailDomainPrefix + obfuscate(domain), nil
}

// Deobfuscate returns the original, unobfuscated email given an
// obfuscated email string originally obtained from Obfuscate.
func Deobfuscate(obfuscatedEmail string) (email string, err error) {
	user, obfuscatedDomainWithPrefix, err := Split(obfuscatedEmail)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(obfuscatedDomainWithPrefix, ObfuscatedEmailDomainPrefix) {
		return "", fmt.Errorf("email domain %q has no obfuscated prefix (%q)", obfuscatedDomainWithPrefix, ObfuscatedEmailDomainPrefix)
	}
	obfuscatedDomain := strings.TrimPrefix(obfuscatedDomainWithPrefix, ObfuscatedEmailDomainPrefix)
	domain, err := deobfuscate(obfuscatedDomain)
	if err != nil {
		return "", err
	}
	return user + "@" + domain, nil
}

// obfuscate obfuscates plaintext.
func obfuscate(plaintext string) string {
	return base64.URLEncoding.EncodeToString(encrypt([]byte(plaintext)))
}

// deobfuscate reverses obfuscation performed by obfuscate.
func deobfuscate(ciphertext string) (string, error) {
	in, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	return string(encrypt(in)), nil
}

func encrypt(in []byte) []byte {
	if in == nil || len(in) == 0 {
		panic("empty `in`")
	}
	aesCipher, err := aes.NewCipher(emailObfuscationKey)
	if err != nil {
		panic("aes.NewCipher: " + err.Error())
	}
	ctr := cipher.NewCTR(aesCipher, emailObfuscationIV)

	ctr.XORKeyStream(in, in)
	return in
}
