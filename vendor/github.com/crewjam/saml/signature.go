package saml

import (
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
)

// Method is part of Signature.
type Method struct {
	Algorithm string `xml:",attr"`
}

// Signature is a model for the Signature object specified by XMLDSIG. This is
// convenience object when constructing XML that you'd like to sign. For example:
//
//    type Foo struct {
//       Stuff string
//       Signature Signature
//    }
//
//    f := Foo{Suff: "hello"}
//    f.Signature = DefaultSignature()
//    buf, _ := xml.Marshal(f)
//    buf, _ = Sign(key, buf)
//
type Signature struct {
	XMLName xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# Signature"`

	CanonicalizationMethod Method             `xml:"SignedInfo>CanonicalizationMethod"`
	SignatureMethod        Method             `xml:"SignedInfo>SignatureMethod"`
	ReferenceTransforms    []Method           `xml:"SignedInfo>Reference>Transforms>Transform"`
	DigestMethod           Method             `xml:"SignedInfo>Reference>DigestMethod"`
	DigestValue            string             `xml:"SignedInfo>Reference>DigestValue"`
	SignatureValue         string             `xml:"SignatureValue"`
	KeyName                string             `xml:"KeyInfo>KeyName,omitempty"`
	X509Certificate        *SignatureX509Data `xml:"KeyInfo>X509Data,omitempty"`
}

// SignatureX509Data represents the <X509Data> element of <Signature>
type SignatureX509Data struct {
	X509Certificate string `xml:"X509Certificate,omitempty"`
}

// DefaultSignature returns a Signature struct that uses the default c14n and SHA1 settings.
func DefaultSignature(pemEncodedPublicKey []byte) Signature {
	// xmlsec wants the key to be base64-encoded but *not* wrapped with the
	// PEM flags
	pemBlock, _ := pem.Decode(pemEncodedPublicKey)
	certStr := base64.StdEncoding.EncodeToString(pemBlock.Bytes)

	return Signature{
		CanonicalizationMethod: Method{
			Algorithm: "http://www.w3.org/TR/2001/REC-xml-c14n-20010315",
		},
		SignatureMethod: Method{
			Algorithm: "http://www.w3.org/2000/09/xmldsig#rsa-sha1",
		},
		ReferenceTransforms: []Method{
			Method{Algorithm: "http://www.w3.org/2000/09/xmldsig#enveloped-signature"},
		},
		DigestMethod: Method{
			Algorithm: "http://www.w3.org/2000/09/xmldsig#sha1",
		},
		X509Certificate: &SignatureX509Data{
			X509Certificate: certStr,
		},
	}
}
