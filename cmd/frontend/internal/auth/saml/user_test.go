package saml

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"reflect"
	"testing"
	"time"

	saml2 "github.com/russellhaering/gosaml2"
	dsig "github.com/russellhaering/goxmldsig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestReadAuthnResponse(t *testing.T) {
	p := &provider{
		samlSP: &saml2.SAMLServiceProvider{
			IdentityProviderSSOURL:      "http://localhost:3220/auth/realms/master",
			IdentityProviderIssuer:      "http://localhost:3220/auth/realms/master",
			Clock:                       dsig.NewFakeClockAt(time.Date(2018, time.May, 20, 17, 12, 6, 0, time.UTC)),
			IDPCertificateStore:         &dsig.MemoryX509CertificateStore{Roots: []*x509.Certificate{idpCert2}},
			SPKeyStore:                  dsig.RandomKeyStoreForTest(),
			AssertionConsumerServiceURL: "http://localhost:3080/.auth/saml/acs",
			ServiceProviderIssuer:       "http://localhost:3080/.auth/saml/metadata",
			AudienceURI:                 "http://localhost:3080/.auth/saml/metadata",
		},
	}
	info, err := readAuthnResponse(p, base64.StdEncoding.EncodeToString([]byte(testAuthnResponse)))
	if err != nil {
		t.Fatal(err)
	}
	info.accountData = nil // skip checking this field
	if want := (&authnResponseInfo{
		spec: extsvc.AccountSpec{
			ServiceType: "saml",
			ServiceID:   "http://localhost:3220/auth/realms/master",
			ClientID:    "http://localhost:3080/.auth/saml/metadata",
			AccountID:   "G-58956f28-7bf5-448d-923a-bd39438c2a9e",
		},
		email:                "bob@example.com",
		unnormalizedUsername: "bob@example.com",
		displayName:          "Bob Yang",
	}); !reflect.DeepEqual(info, want) {
		t.Errorf("got != want\n got %+v\nwant %+v", info, want)
	}
}

func TestReadAuthnResponseWithUsernameKey(t *testing.T) {
	p := &provider{
		config: schema.SAMLAuthProvider{
			UsernameAttributeNames: []string{"givenName"},
		},
		samlSP: &saml2.SAMLServiceProvider{
			IdentityProviderSSOURL:      "http://localhost:3220/auth/realms/master",
			IdentityProviderIssuer:      "http://localhost:3220/auth/realms/master",
			Clock:                       dsig.NewFakeClockAt(time.Date(2018, time.May, 20, 17, 12, 6, 0, time.UTC)),
			IDPCertificateStore:         &dsig.MemoryX509CertificateStore{Roots: []*x509.Certificate{idpCert2}},
			SPKeyStore:                  dsig.RandomKeyStoreForTest(),
			AssertionConsumerServiceURL: "http://localhost:3080/.auth/saml/acs",
			ServiceProviderIssuer:       "http://localhost:3080/.auth/saml/metadata",
			AudienceURI:                 "http://localhost:3080/.auth/saml/metadata",
		},
	}
	info, err := readAuthnResponse(p, base64.StdEncoding.EncodeToString([]byte(testAuthnResponse)))
	if err != nil {
		t.Fatal(err)
	}
	info.accountData = nil // skip checking this field
	if want := (&authnResponseInfo{
		spec: extsvc.AccountSpec{
			ServiceType: "saml",
			ServiceID:   "http://localhost:3220/auth/realms/master",
			ClientID:    "http://localhost:3080/.auth/saml/metadata",
			AccountID:   "G-58956f28-7bf5-448d-923a-bd39438c2a9e",
		},
		email:                "bob@example.com",
		unnormalizedUsername: "Bob",
		displayName:          "Bob Yang",
	}); !reflect.DeepEqual(info, want) {
		t.Errorf("got != want\n got %+v\nwant %+v", info, want)
	}
}

func TestReadAuthnResponseWithUsernameKeyMultipleCerts(t *testing.T) {
	p := &provider{
		config: schema.SAMLAuthProvider{
			UsernameAttributeNames: []string{"givenName"},
		},
		samlSP: &saml2.SAMLServiceProvider{
			IdentityProviderSSOURL:      "http://localhost:3220/auth/realms/master",
			IdentityProviderIssuer:      "http://localhost:3220/auth/realms/master",
			Clock:                       dsig.NewFakeClockAt(time.Date(2018, time.May, 20, 17, 12, 6, 0, time.UTC)),
			IDPCertificateStore:         &dsig.MemoryX509CertificateStore{Roots: []*x509.Certificate{idpCert1, idpCert2}},
			SPKeyStore:                  dsig.RandomKeyStoreForTest(),
			AssertionConsumerServiceURL: "http://localhost:3080/.auth/saml/acs",
			ServiceProviderIssuer:       "http://localhost:3080/.auth/saml/metadata",
			AudienceURI:                 "http://localhost:3080/.auth/saml/metadata",
		},
	}

	// idpCert1
	info, err := readAuthnResponse(p, base64.StdEncoding.EncodeToString([]byte(testAuthnResponse2)))
	if err != nil {
		t.Fatal(err)
	}
	info.accountData = nil // skip checking this field
	if want := (&authnResponseInfo{
		spec: extsvc.AccountSpec{
			ServiceType: "saml",
			ServiceID:   "http://localhost:3220/auth/realms/master",
			ClientID:    "http://localhost:3080/.auth/saml/metadata",
			AccountID:   "G-58956f28-7bf5-448d-923a-bd39438c2a9e",
		},
		email:                "bob@example.com",
		unnormalizedUsername: "Bob",
		displayName:          "Bob Yang",
	}); !reflect.DeepEqual(info, want) {
		t.Errorf("got != want\n got %+v\nwant %+v", info, want)
	}

	// idpCert2
	info, err = readAuthnResponse(p, base64.StdEncoding.EncodeToString([]byte(testAuthnResponse)))
	if err != nil {
		t.Fatal(err)
	}
	info.accountData = nil // skip checking this field
	if want := (&authnResponseInfo{
		spec: extsvc.AccountSpec{
			ServiceType: "saml",
			ServiceID:   "http://localhost:3220/auth/realms/master",
			ClientID:    "http://localhost:3080/.auth/saml/metadata",
			AccountID:   "G-58956f28-7bf5-448d-923a-bd39438c2a9e",
		},
		email:                "bob@example.com",
		unnormalizedUsername: "Bob",
		displayName:          "Bob Yang",
	}); !reflect.DeepEqual(info, want) {
		t.Errorf("got != want\n got %+v\nwant %+v", info, want)
	}
}

var idpCert1 = func() *x509.Certificate {
	b, _ := pem.Decode([]byte(`-----BEGIN CERTIFICATE-----
MIIC0TCCAbmgAwIBAgIUUU1HCBeiRgF40N94DKCpUm+AFEIwDQYJKoZIhvcNAQEL
BQAwETEPMA0GA1UEAwwGbWFzdGVyMB4XDTE3MDYxNDEzMjA1M1oXDTI3MDYxMjEz
MjA1M1owETEPMA0GA1UEAwwGbWFzdGVyMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEAuec6f7U/8AjPiqrHdaWyBqdTzcIAw+31vD1b8+Ycj3fYbBO6G33d
KrJdYDS2U7Cpd/N7/1hGSt3VYNfm718xiRjzzq+sXvBZx4J3QG3LEe9BLy7Lx0fM
BXdX2xbO+br1Te4/EOus5rl6GarxS0lSuKr8t8QfOY+thTWKYYqxlhZMPaI5WpIC
8O4OuSBiIev82aEHg+WG6h6oS0WHFFwZF10okMR3oqIfPZk4M2Cb+SA6J2rAd8sn
rx3AF5r/EOnZUpDqV+UIaNfnGwCXAwA/+nVSJlhovLTA0s54dLXcqg6iAk/8UxZQ
s5cg3KondIMRAtrRVtHRH24ekDEHz95OlwIDAQABoyEwHzAdBgNVHQ4EFgQUyaPW
FfkOESSnbQknlzB2lBpInGowDQYJKoZIhvcNAQELBQADggEBAKWj5crOgMIeKjcv
oxWeDz5EZccBq810EAEHcta03M5dnfXokvzTxdOMQ3quPCEQkmK3g4A5wDrKPlag
aA3egz3Z14ScIWBRHWaiEtYD5ALt4k/rIapnyJoANRjbPmNglabkk7TfzzqB5pQ5
r6tqe5OQx3bR85zQr/Xkj1dTKRdc49efKfmz5be32H1l2h3DHGbqWLHI4jEL8Xv5
xbLyV88tz5Eah90nV1KfrTQdXDjec0o1RdArgV03wgVmh+8OUp/RgN3xVS8BMrFU
YN8bOpm1XA8Txs3u0ESLR2ztDpVcY8hKGE6ZnuHpKg2PEhx2vBI01oWf6PPyB/Sn
UFn0dMY=
-----END CERTIFICATE-----`))
	c, _ := x509.ParseCertificate(b.Bytes)
	return c
}()

var idpCert2 = func() *x509.Certificate {
	b, _ := pem.Decode([]byte(`-----BEGIN CERTIFICATE-----
MIICmzCCAYMCBgFjcZU/LjANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDDAZtYXN0ZXIwHhcNMTgwNTE4MDQ0ODE2WhcNMjgwNTE4MDQ0OTU2WjARMQ8wDQYDVQQDDAZtYXN0ZXIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDXZpJeHraEt9FPk478+RoMtP9RV83Ew/XRZhNKI4BPoY5MjRVuvaabvMOE5X1AK9Z0cEU++m/Y0LuHg3A4kQdPw3BGPBfGm0WSD6DEN42TcF3dc8XBA/osDNW5i6rZM071che8XtKNHcW9ZAv9ETfJeUb4NHFRkRg3K1lZ5kCwt0JNo+0akQ2EdQXXu/uEeQV49rOADr+Lp6GLhmGeCckC8xzBiNxZwR4pJsz9XWgB6fSdpIGvWhAnBfFZyyZIHnVuRnm2wJ53Exg6h2RB3SFYu3PXXuIHeuH71pel5WwnecTVTwV/RMwkAGLdCNC9jp9tdDtThhWLn4E9D0wZkpU9AgMBAAEwDQYJKoZIhvcNAQELBQADggEBAKT/zyjvSM09Fk2ON4rMSExnyrw6LXuJJOZlB0eD22KruQ53AikfKz5nJLCFLc0PT4PmK06s9OF0HG95k4jiiuvAdNMXZSLUGNcbaODeJ/ZzCJJp0cB2rWEmAqbKruXzBpTFttlgsW4mgpkvGxORztfhksiyAX0bLcNWtsQecl3fpvoVrJiIHXStD3c/v4exE2QPkuvhLCzwI2oXrrhrovyTKjCbyn2//lqOfFziA8X/ini3R/L4UzTVB5SWAz/LtkpgipPOwNpVqwErnZamexm6S38QX+OZ+uhZY/1JfTugs9vpXwRvj/xamGr8r+MqornuQiEBBNiCbCJ6B4iUWh4=
-----END CERTIFICATE-----`))
	c, _ := x509.ParseCertificate(b.Bytes)
	return c
}()

const (
	testAuthnResponse = `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" Destination="http://localhost:3080/.auth/saml/acs" ID="ID_4f2db416-8815-4d9c-84b5-871eb79ce4f4" InResponseTo="_b5744ff7-7066-4db6-a19f-baa925227c8a" IssueInstant="2018-05-20T17:12:06.795Z" Version="2.0"><saml:Issuer>http://localhost:3220/auth/realms/master</saml:Issuer><dsig:Signature xmlns:dsig="http://www.w3.org/2000/09/xmldsig#"><dsig:SignedInfo><dsig:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/><dsig:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"/><dsig:Reference URI="#ID_4f2db416-8815-4d9c-84b5-871eb79ce4f4"><dsig:Transforms><dsig:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/><dsig:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/></dsig:Transforms><dsig:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"/><dsig:DigestValue>LAWFTwzvNHyOLqwimF3QR5dJgEfCxs2RXUdp+raMdRk=</dsig:DigestValue></dsig:Reference></dsig:SignedInfo><dsig:SignatureValue>xg13DUl1G80iCxATDN0R2QGUBHq+n6N9J389zTBM36ploAbWtvnI29IuW+aRaO69cUKsHBGH3YIV7njUNcDOOHMX1b9K+hooqaRyGfKISnvnaLZ+/R3yXZf+pAFshvtgWkaS+29zmNP9+5j3X/j9Gj9buoIlL5f51MO8fXlYJtdxqIhFoYZWcrttstxQhENjskFezYPyepl5F49m+FY5nYKh75WcG51NI+/VSYqWQd7MeUompPTONbt8Kwtj7YGizNbJseEOt1EI5wn+7eFvq/DkpJAuKDB4jnjbjadQmEbUIfKew5u/EEn6WDVnidL9vQQh/ZVOmFqL77iqBbQPLw==</dsig:SignatureValue><dsig:KeyInfo><dsig:KeyName>jR3UQTQOE9k8iqTK77NrOBahhyFNT2p3B2lF1I3ov1g</dsig:KeyName><dsig:X509Data><dsig:X509Certificate>MIICmzCCAYMCBgFjcZU/LjANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDDAZtYXN0ZXIwHhcNMTgwNTE4MDQ0ODE2WhcNMjgwNTE4MDQ0OTU2WjARMQ8wDQYDVQQDDAZtYXN0ZXIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDXZpJeHraEt9FPk478+RoMtP9RV83Ew/XRZhNKI4BPoY5MjRVuvaabvMOE5X1AK9Z0cEU++m/Y0LuHg3A4kQdPw3BGPBfGm0WSD6DEN42TcF3dc8XBA/osDNW5i6rZM071che8XtKNHcW9ZAv9ETfJeUb4NHFRkRg3K1lZ5kCwt0JNo+0akQ2EdQXXu/uEeQV49rOADr+Lp6GLhmGeCckC8xzBiNxZwR4pJsz9XWgB6fSdpIGvWhAnBfFZyyZIHnVuRnm2wJ53Exg6h2RB3SFYu3PXXuIHeuH71pel5WwnecTVTwV/RMwkAGLdCNC9jp9tdDtThhWLn4E9D0wZkpU9AgMBAAEwDQYJKoZIhvcNAQELBQADggEBAKT/zyjvSM09Fk2ON4rMSExnyrw6LXuJJOZlB0eD22KruQ53AikfKz5nJLCFLc0PT4PmK06s9OF0HG95k4jiiuvAdNMXZSLUGNcbaODeJ/ZzCJJp0cB2rWEmAqbKruXzBpTFttlgsW4mgpkvGxORztfhksiyAX0bLcNWtsQecl3fpvoVrJiIHXStD3c/v4exE2QPkuvhLCzwI2oXrrhrovyTKjCbyn2//lqOfFziA8X/ini3R/L4UzTVB5SWAz/LtkpgipPOwNpVqwErnZamexm6S38QX+OZ+uhZY/1JfTugs9vpXwRvj/xamGr8r+MqornuQiEBBNiCbCJ6B4iUWh4=</dsig:X509Certificate></dsig:X509Data><dsig:KeyValue><dsig:RSAKeyValue><dsig:Modulus>12aSXh62hLfRT5OO/PkaDLT/UVfNxMP10WYTSiOAT6GOTI0Vbr2mm7zDhOV9QCvWdHBFPvpv2NC7h4NwOJEHT8NwRjwXxptFkg+gxDeNk3Bd3XPFwQP6LAzVuYuq2TNO9XIXvF7SjR3FvWQL/RE3yXlG+DRxUZEYNytZWeZAsLdCTaPtGpENhHUF17v7hHkFePazgA6/i6ehi4ZhngnJAvMcwYjcWcEeKSbM/V1oAen0naSBr1oQJwXxWcsmSB51bkZ5tsCedxMYOodkQd0hWLtz117iB3rh+9aXpeVsJ3nE1U8Ff0TMJABi3QjQvY6fbXQ7U4YVi5+BPQ9MGZKVPQ==</dsig:Modulus><dsig:Exponent>AQAB</dsig:Exponent></dsig:RSAKeyValue></dsig:KeyValue></dsig:KeyInfo></dsig:Signature><samlp:Status><samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/></samlp:Status><saml:Assertion xmlns="urn:oasis:names:tc:SAML:2.0:assertion" ID="ID_68f3f1bf-05b3-4c13-a2c6-63a4885ed70b" IssueInstant="2018-05-20T17:12:06.795Z" Version="2.0"><saml:Issuer>http://localhost:3220/auth/realms/master</saml:Issuer><saml:Subject><saml:NameID Format="urn:oasis:names:tc:SAML:2.0:nameid-format:persistent">G-58956f28-7bf5-448d-923a-bd39438c2a9e</saml:NameID><saml:SubjectConfirmation Method="urn:oasis:names:tc:SAML:2.0:cm:bearer"><saml:SubjectConfirmationData InResponseTo="_b5744ff7-7066-4db6-a19f-baa925227c8a" NotOnOrAfter="2018-05-20T17:13:04.795Z" Recipient="http://localhost:3080/.auth/saml/acs"/></saml:SubjectConfirmation></saml:Subject><saml:Conditions NotBefore="2018-05-20T17:12:04.795Z" NotOnOrAfter="2018-05-20T17:13:04.795Z"><saml:AudienceRestriction><saml:Audience>http://localhost:3080/.auth/saml/metadata</saml:Audience></saml:AudienceRestriction></saml:Conditions><saml:AuthnStatement AuthnInstant="2018-05-20T17:12:06.795Z" SessionIndex="0c9f6960-c426-4d45-9b0c-b11870cd8338::bc174b26-a300-4e78-9d76-49f2521e7b65"><saml:AuthnContext><saml:AuthnContextClassRef>urn:oasis:names:tc:SAML:2.0:ac:classes:unspecified</saml:AuthnContextClassRef></saml:AuthnContext></saml:AuthnStatement><saml:AttributeStatement><saml:Attribute FriendlyName="surname" Name="urn:oid:2.5.4.4" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">Yang</saml:AttributeValue></saml:Attribute><saml:Attribute FriendlyName="givenName" Name="urn:oid:2.5.4.42" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">Bob</saml:AttributeValue></saml:Attribute><saml:Attribute FriendlyName="email" Name="urn:oid:1.2.840.113549.1.9.1" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">bob@example.com</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">view-profile</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">uma_authorization</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">manage-account</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">manage-account-links</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">admin</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">view-realm</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">manage-identity-providers</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">query-realms</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">manage-clients</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">query-clients</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">manage-authorization</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">query-users</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">view-users</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">manage-users</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">view-events</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">create-realm</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">create-client</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">view-clients</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">view-identity-providers</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">query-groups</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">manage-events</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">impersonation</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">view-authorization</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">manage-realm</saml:AttributeValue></saml:Attribute></saml:AttributeStatement></saml:Assertion></samlp:Response>`
	// Request signed using https://www.samltool.com/sign_response.php
	testAuthnResponse2 = `<?xml version="1.0"?>
<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" Destination="http://localhost:3080/.auth/saml/acs" ID="pfx4be74087-5efa-c285-513e-ecc4c1b11717" InResponseTo="_b5744ff7-7066-4db6-a19f-baa925227c8a" IssueInstant="2018-05-20T17:12:06.795Z" Version="2.0"><saml:Issuer>http://localhost:3220/auth/realms/master</saml:Issuer><ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
  <ds:SignedInfo><ds:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
    <ds:SignatureMethod Algorithm="http://www.w3.org/2000/09/xmldsig#rsa-sha1"/>
  <ds:Reference URI="#pfx4be74087-5efa-c285-513e-ecc4c1b11717"><ds:Transforms><ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/><ds:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/></ds:Transforms><ds:DigestMethod Algorithm="http://www.w3.org/2000/09/xmldsig#sha1"/><ds:DigestValue>UbLzMnLBvvaZRpi4stMVcaKrJTo=</ds:DigestValue></ds:Reference></ds:SignedInfo><ds:SignatureValue>S9j4KhpH1qZM+3SYcqvlPFMIJHTJ2pP44RHpI20q/6pXsPv69wFV3JJ9iczANDWwwLmo29t8fTMJxbut/BcpaIj8Gf/d+CunNblgWQLx8Z162bHoTpqyjOTjF2IWyL7TP0nCyqu7RbBZ+HepWZTKa3PeWp7KGlGsFrFZVh1Qv6xjjl1tlcbKDxqenIdCxt0z4C0tX+k6ZEnWIvjt2Ml1Wtf+aXDDMc2N9Xgq7F9M+PBzO+miQ1HEdnYvPEKQIz/0qL3P53rQ34xObysY5thBenkFJsHEkeqIs5wjA5mG7ALLugEOLtoetB8frCKCaaL4xe7+0ElYy9zCGTYBk1cXdw==</ds:SignatureValue>
<ds:KeyInfo><ds:X509Data><ds:X509Certificate>MIIC0TCCAbmgAwIBAgIUUU1HCBeiRgF40N94DKCpUm+AFEIwDQYJKoZIhvcNAQELBQAwETEPMA0GA1UEAwwGbWFzdGVyMB4XDTE3MDYxNDEzMjA1M1oXDTI3MDYxMjEzMjA1M1owETEPMA0GA1UEAwwGbWFzdGVyMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuec6f7U/8AjPiqrHdaWyBqdTzcIAw+31vD1b8+Ycj3fYbBO6G33dKrJdYDS2U7Cpd/N7/1hGSt3VYNfm718xiRjzzq+sXvBZx4J3QG3LEe9BLy7Lx0fMBXdX2xbO+br1Te4/EOus5rl6GarxS0lSuKr8t8QfOY+thTWKYYqxlhZMPaI5WpIC8O4OuSBiIev82aEHg+WG6h6oS0WHFFwZF10okMR3oqIfPZk4M2Cb+SA6J2rAd8snrx3AF5r/EOnZUpDqV+UIaNfnGwCXAwA/+nVSJlhovLTA0s54dLXcqg6iAk/8UxZQs5cg3KondIMRAtrRVtHRH24ekDEHz95OlwIDAQABoyEwHzAdBgNVHQ4EFgQUyaPWFfkOESSnbQknlzB2lBpInGowDQYJKoZIhvcNAQELBQADggEBAKWj5crOgMIeKjcvoxWeDz5EZccBq810EAEHcta03M5dnfXokvzTxdOMQ3quPCEQkmK3g4A5wDrKPlagaA3egz3Z14ScIWBRHWaiEtYD5ALt4k/rIapnyJoANRjbPmNglabkk7TfzzqB5pQ5r6tqe5OQx3bR85zQr/Xkj1dTKRdc49efKfmz5be32H1l2h3DHGbqWLHI4jEL8Xv5xbLyV88tz5Eah90nV1KfrTQdXDjec0o1RdArgV03wgVmh+8OUp/RgN3xVS8BMrFUYN8bOpm1XA8Txs3u0ESLR2ztDpVcY8hKGE6ZnuHpKg2PEhx2vBI01oWf6PPyB/SnUFn0dMY=</ds:X509Certificate></ds:X509Data></ds:KeyInfo></ds:Signature><dsig:Signature xmlns:dsig="http://www.w3.org/2000/09/xmldsig#"><dsig:SignedInfo><dsig:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/><dsig:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"/><dsig:Reference URI="#ID_4f2db416-8815-4d9c-84b5-871eb79ce4f4"><dsig:Transforms><dsig:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/><dsig:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/></dsig:Transforms><dsig:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"/><dsig:DigestValue>LAWFTwzvNHyOLqwimF3QR5dJgEfCxs2RXUdp+raMdRk=</dsig:DigestValue></dsig:Reference></dsig:SignedInfo><dsig:SignatureValue>xg13DUl1G80iCxATDN0R2QGUBHq+n6N9J389zTBM36ploAbWtvnI29IuW+aRaO69cUKsHBGH3YIV7njUNcDOOHMX1b9K+hooqaRyGfKISnvnaLZ+/R3yXZf+pAFshvtgWkaS+29zmNP9+5j3X/j9Gj9buoIlL5f51MO8fXlYJtdxqIhFoYZWcrttstxQhENjskFezYPyepl5F49m+FY5nYKh75WcG51NI+/VSYqWQd7MeUompPTONbt8Kwtj7YGizNbJseEOt1EI5wn+7eFvq/DkpJAuKDB4jnjbjadQmEbUIfKew5u/EEn6WDVnidL9vQQh/ZVOmFqL77iqBbQPLw==</dsig:SignatureValue><dsig:KeyInfo><dsig:KeyName>jR3UQTQOE9k8iqTK77NrOBahhyFNT2p3B2lF1I3ov1g</dsig:KeyName><dsig:X509Data><dsig:X509Certificate>MIICmzCCAYMCBgFjcZU/LjANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDDAZtYXN0ZXIwHhcNMTgwNTE4MDQ0ODE2WhcNMjgwNTE4MDQ0OTU2WjARMQ8wDQYDVQQDDAZtYXN0ZXIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDXZpJeHraEt9FPk478+RoMtP9RV83Ew/XRZhNKI4BPoY5MjRVuvaabvMOE5X1AK9Z0cEU++m/Y0LuHg3A4kQdPw3BGPBfGm0WSD6DEN42TcF3dc8XBA/osDNW5i6rZM071che8XtKNHcW9ZAv9ETfJeUb4NHFRkRg3K1lZ5kCwt0JNo+0akQ2EdQXXu/uEeQV49rOADr+Lp6GLhmGeCckC8xzBiNxZwR4pJsz9XWgB6fSdpIGvWhAnBfFZyyZIHnVuRnm2wJ53Exg6h2RB3SFYu3PXXuIHeuH71pel5WwnecTVTwV/RMwkAGLdCNC9jp9tdDtThhWLn4E9D0wZkpU9AgMBAAEwDQYJKoZIhvcNAQELBQADggEBAKT/zyjvSM09Fk2ON4rMSExnyrw6LXuJJOZlB0eD22KruQ53AikfKz5nJLCFLc0PT4PmK06s9OF0HG95k4jiiuvAdNMXZSLUGNcbaODeJ/ZzCJJp0cB2rWEmAqbKruXzBpTFttlgsW4mgpkvGxORztfhksiyAX0bLcNWtsQecl3fpvoVrJiIHXStD3c/v4exE2QPkuvhLCzwI2oXrrhrovyTKjCbyn2//lqOfFziA8X/ini3R/L4UzTVB5SWAz/LtkpgipPOwNpVqwErnZamexm6S38QX+OZ+uhZY/1JfTugs9vpXwRvj/xamGr8r+MqornuQiEBBNiCbCJ6B4iUWh4=</dsig:X509Certificate></dsig:X509Data><dsig:KeyValue><dsig:RSAKeyValue><dsig:Modulus>12aSXh62hLfRT5OO/PkaDLT/UVfNxMP10WYTSiOAT6GOTI0Vbr2mm7zDhOV9QCvWdHBFPvpv2NC7h4NwOJEHT8NwRjwXxptFkg+gxDeNk3Bd3XPFwQP6LAzVuYuq2TNO9XIXvF7SjR3FvWQL/RE3yXlG+DRxUZEYNytZWeZAsLdCTaPtGpENhHUF17v7hHkFePazgA6/i6ehi4ZhngnJAvMcwYjcWcEeKSbM/V1oAen0naSBr1oQJwXxWcsmSB51bkZ5tsCedxMYOodkQd0hWLtz117iB3rh+9aXpeVsJ3nE1U8Ff0TMJABi3QjQvY6fbXQ7U4YVi5+BPQ9MGZKVPQ==</dsig:Modulus><dsig:Exponent>AQAB</dsig:Exponent></dsig:RSAKeyValue></dsig:KeyValue></dsig:KeyInfo></dsig:Signature><samlp:Status><samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/></samlp:Status><saml:Assertion xmlns="urn:oasis:names:tc:SAML:2.0:assertion" ID="ID_68f3f1bf-05b3-4c13-a2c6-63a4885ed70b" IssueInstant="2018-05-20T17:12:06.795Z" Version="2.0"><saml:Issuer>http://localhost:3220/auth/realms/master</saml:Issuer><saml:Subject><saml:NameID Format="urn:oasis:names:tc:SAML:2.0:nameid-format:persistent">G-58956f28-7bf5-448d-923a-bd39438c2a9e</saml:NameID><saml:SubjectConfirmation Method="urn:oasis:names:tc:SAML:2.0:cm:bearer"><saml:SubjectConfirmationData InResponseTo="_b5744ff7-7066-4db6-a19f-baa925227c8a" NotOnOrAfter="2018-05-20T17:13:04.795Z" Recipient="http://localhost:3080/.auth/saml/acs"/></saml:SubjectConfirmation></saml:Subject><saml:Conditions NotBefore="2018-05-20T17:12:04.795Z" NotOnOrAfter="2018-05-20T17:13:04.795Z"><saml:AudienceRestriction><saml:Audience>http://localhost:3080/.auth/saml/metadata</saml:Audience></saml:AudienceRestriction></saml:Conditions><saml:AuthnStatement AuthnInstant="2018-05-20T17:12:06.795Z" SessionIndex="0c9f6960-c426-4d45-9b0c-b11870cd8338::bc174b26-a300-4e78-9d76-49f2521e7b65"><saml:AuthnContext><saml:AuthnContextClassRef>urn:oasis:names:tc:SAML:2.0:ac:classes:unspecified</saml:AuthnContextClassRef></saml:AuthnContext></saml:AuthnStatement><saml:AttributeStatement><saml:Attribute FriendlyName="surname" Name="urn:oid:2.5.4.4" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">Yang</saml:AttributeValue></saml:Attribute><saml:Attribute FriendlyName="givenName" Name="urn:oid:2.5.4.42" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">Bob</saml:AttributeValue></saml:Attribute><saml:Attribute FriendlyName="email" Name="urn:oid:1.2.840.113549.1.9.1" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">bob@example.com</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">view-profile</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">uma_authorization</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">manage-account</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">manage-account-links</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">admin</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">view-realm</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">manage-identity-providers</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">query-realms</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">manage-clients</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">query-clients</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">manage-authorization</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">query-users</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">view-users</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">manage-users</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">view-events</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">create-realm</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">create-client</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">view-clients</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">view-identity-providers</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">query-groups</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">manage-events</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">impersonation</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">view-authorization</saml:AttributeValue></saml:Attribute><saml:Attribute Name="Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic"><saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">manage-realm</saml:AttributeValue></saml:Attribute></saml:AttributeStatement></saml:Assertion></samlp:Response>`
)

func TestGetPublicExternalAccountData(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected *extsvc.PublicAccountData
		err      string
	}{
		{
			name:     "No values",
			input:    "",
			expected: nil,
			err:      "could not find data for the external account",
		},
		{
			name: "Empty values",
			input: `{
				"Values": {}
			}`,
			expected: nil,
		},
		{
			name: "Case insensitive name",
			input: `{
				"Values": {
					"Username": {
						"Values": [{ "Value": "Alice" }]
					},
					"Email": {
						"Values": [{ "Value": "alice@acme.com" }]
					}
				}
			}`,
			expected: &extsvc.PublicAccountData{
				DisplayName: "Alice",
			},
		},
		{
			name: "schema attributes with URL",
			input: `{
				"Values": {
					"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress": {
						"Values": [{ "Value": "alice@acme.com" }]
					}
				}
			}`,
			expected: &extsvc.PublicAccountData{
				DisplayName: "alice@acme.com",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := extsvc.AccountData{}
			if tc.input == "" {
				data.Data = nil
			} else {
				data.Data = extsvc.NewUnencryptedData([]byte(tc.input))
			}

			publicData, err := GetPublicExternalAccountData(context.Background(), &data)
			if tc.err != "" {
				require.EqualError(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expected, publicData)
		})
	}
}
