package invite

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mattbaird/gochimp"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/notif"
)

type TokenPayload struct {
	OrgID   int32
	OrgName string
	Email   string
}

func getSecretKey() ([]byte, error) {
	encoded := conf.Get().SecretKey
	if encoded == "" {
		return nil, errors.New("secret key is not set in site config")
	}
	v, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errors.New("error base64-decoding secret key")
	}
	return v, err
}

func ParseToken(tokenString string) (*TokenPayload, error) {
	payload, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return 0, fmt.Errorf("invite: unexpected signing method %v", token.Header["alg"])
		}
		return getSecretKey()
	})
	if err != nil {
		return nil, err
	}
	claims, ok := payload.Claims.(jwt.MapClaims)
	if !ok || !payload.Valid {
		return nil, errors.New("invite: invalid token")
	}

	orgID, ok := claims["orgID"].(float64)
	if !ok {
		return nil, errors.New("invite: unexpected org id")
	}
	orgName, ok := claims["orgName"].(string)
	if !ok {
		return nil, errors.New("invite: unexpected org name")
	}
	email, ok := claims["email"].(string)
	if !ok {
		return nil, errors.New("invite: unexpected email")
	}

	return &TokenPayload{OrgID: int32(orgID), OrgName: orgName, Email: email}, nil
}

func CreateOrgToken(email string, org *sourcegraph.Org) (string, error) {
	payload := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":   email,
		"orgID":   org.ID,
		"orgName": org.Name, // So the accept invite UI can display the name of the org
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	})
	key, err := getSecretKey()
	if err != nil {
		return "", err
	}
	return payload.SignedString(key)
}

func SendEmail(inviteEmail, fromName, orgName, inviteURL string) {
	config := &notif.EmailConfig{
		Template:  "invite-user",
		FromName:  fromName,
		FromEmail: "noreply@sourcegraph.com",
		ToEmail:   inviteEmail,
		Subject:   fmt.Sprintf("%s has invited you to join %s on Sourcegraph", fromName, orgName),
	}

	notif.SendMandrillTemplate(config, []gochimp.Var{}, []gochimp.Var{
		gochimp.Var{Name: "INVITE_URL", Content: inviteURL},
		gochimp.Var{Name: "ORG", Content: orgName},
		gochimp.Var{Name: "FROM_USER", Content: fromName},
	})
}
