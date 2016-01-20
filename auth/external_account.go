package auth

import (
	"fmt"
	"os"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/platform/storage"
)

// ExtAccountsStore contains methods for storing information about
// Sourcegraph user accounts linked to external accounts.
type ExtAccountsStore struct{}

const accountsBucket = "users"

type ExtAccount struct {
	Host     string
	UID      int32
	ExtLogin string
}

const orgsBucket = "orgs"

type ExtOrg struct {
	Host  string
	Org   string
	Users []int32
}

func (s *ExtAccountsStore) storage(ctx context.Context) storage.System {
	return storage.Namespace(ctx, "core.ext-accounts", "")
}

func (s *ExtAccountsStore) getAccountKey(host string, uid int32) string {
	return fmt.Sprintf("%s:%d", host, uid)
}

func (s *ExtAccountsStore) Set(ctx context.Context, host string, uid int32, extLogin string) error {
	fs := s.storage(ctx)
	return storage.PutJSON(fs, accountsBucket, s.getAccountKey(host, uid), ExtAccount{
		Host:     host,
		UID:      uid,
		ExtLogin: extLogin,
	})
}

func (s *ExtAccountsStore) Get(ctx context.Context, host string, uid int32) (ExtAccount, error) {
	acct := ExtAccount{}
	fs := s.storage(ctx)
	err := storage.GetJSON(fs, accountsBucket, s.getAccountKey(host, uid), &acct)
	if err != nil {
		if os.IsNotExist(err) {
			return ExtAccount{}, nil
		}
		return ExtAccount{}, err
	}

	return acct, nil
}

func (s *ExtAccountsStore) getOrgKey(host, org string) string {
	return fmt.Sprintf("%s:%s", host, org)
}

func (s *ExtAccountsStore) Append(ctx context.Context, host, org string, uid int32) error {
	fs := s.storage(ctx)

	extOrg, err := s.GetAll(ctx, host, org)
	if err != nil {
		return err
	}
	extOrg.Users = append(extOrg.Users, uid)

	return storage.PutJSON(fs, orgsBucket, s.getOrgKey(host, org), extOrg)
}

func (s *ExtAccountsStore) GetAll(ctx context.Context, host, org string) (ExtOrg, error) {
	extOrg := ExtOrg{}
	fs := s.storage(ctx)
	err := storage.GetJSON(fs, orgsBucket, s.getOrgKey(host, org), &extOrg)
	if err != nil {
		if os.IsNotExist(err) {
			return ExtOrg{Host: host, Org: org, Users: make([]int32, 0)}, nil
		}
		return ExtOrg{Host: host, Org: org, Users: make([]int32, 0)}, err
	}

	return extOrg, nil
}
