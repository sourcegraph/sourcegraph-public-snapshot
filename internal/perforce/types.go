package perforce

import (
	v1 "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
)

type User struct {
	Username string
	Email    string
}

func UserFromProto(proto *v1.PerforceUser) *User {
	return &User{
		Username: proto.GetUsername(),
		Email:    proto.GetEmail(),
	}
}

func (u *User) ToProto() *v1.PerforceUser {
	return &v1.PerforceUser{
		Username: u.Username,
		Email:    u.Email,
	}
}

type Protect struct {
	Level       string
	EntityType  string
	EntityName  string
	Match       string
	IsExclusion bool
	Host        string
}

func ProtectFromProto(proto *v1.PerforceProtect) *Protect {
	return &Protect{
		Level:       proto.GetLevel(),
		EntityType:  proto.GetEntityType(),
		EntityName:  proto.GetEntityName(),
		Match:       proto.GetMatch(),
		IsExclusion: proto.GetIsExclusion(),
		Host:        proto.GetHost(),
	}
}

func (p *Protect) ToProto() *v1.PerforceProtect {
	return &v1.PerforceProtect{
		Level:       p.Level,
		EntityType:  p.EntityType,
		EntityName:  p.EntityName,
		Match:       p.Match,
		IsExclusion: p.IsExclusion,
		Host:        p.Host,
	}
}
