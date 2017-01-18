package graphqlbackend

import (
	"crypto/md5"
	"fmt"
	"strings"
)

type personResolver struct {
	name         string
	email        string
	gravatarHash string
}

func (r *personResolver) Name() string {
	return r.name
}

func (r *personResolver) Email() string {
	return r.email
}

func (r *personResolver) GravatarHash() string {
	if r.email != "" {
		h := md5.New()
		h.Write([]byte(strings.ToLower(r.email)))
		return fmt.Sprintf("%x", h.Sum(nil))
	}

	return ""
}
