package database

import (
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	errorCodeUserWithEmailExists          = "err_user_with_such_email_exists"
	errorCodeAccessRequestWithEmailExists = "err_access_request_with_such_email_exists"
)

// errCannotCreateAccessRequest is the error that is returned when a request_access cannot be added to the DB due to a constraint.
type errCannotCreateAccessRequest struct {
	code string
}

func (err errCannotCreateAccessRequest) Error() string {
	return fmt.Sprintf("cannot create user: %v", err.code)
}

// IsAccessRequestUserWithEmailExists reports whether err is an error indicating that the access request email was already taken by a signed in user.
func IsAccessRequestUserWithEmailExists(err error) bool {
	var e errCannotCreateAccessRequest
	return errors.As(err, &e) && e.code == errorCodeUserWithEmailExists
}

// IsAccessRequestWithEmailExists reports whether err is an error indicating that the access request was already created.
func IsAccessRequestWithEmailExists(err error) bool {
	var e errCannotCreateAccessRequest
	return errors.As(err, &e) && e.code == errorCodeAccessRequestWithEmailExists
}

// UserStore provides access to the `users` table.
//
// For a detailed overview of the schema, see schema.md.
type AccessRequestStore interface {
	basestore.ShareableStore
	Count(context.Context, *UsersListOptions) (int, error)
	Create(context.Context, NewAccessRequest) (*types.AccessRequest, error)
	Delete(context.Context, int32) error
	GetByID(context.Context, int32) (*types.User, error)
	GetByEmail(context.Context, string) (*types.User, error)
	HardDelete(context.Context, int32) error
	List(context.Context, *UsersListOptions) (_ []*types.User, err error)
	Update(context.Context, int32, UserUpdate) error
}

type accessRequestStore struct {
	*basestore.Store
	logger log.Logger
}

type NewAccessRequest struct {
	Name string
	Email string
	AdditionalInfo string
}

// AccessRequestsWith instantiates and returns a new accessRequestStore using the other store handle.
func AccessRequestsWith(other basestore.ShareableStore, logger log.Logger) AccessRequestStore {
	return &accessRequestStore{Store: basestore.NewWithHandle(other.Handle()), logger: logger}
}

const accessRequestCreateQueryFmtStr = `
INSERT INTO
	access_requests (name, email, additional_info)
VALUES ( %s, %s, %s )
RETURNING id, name, email, additional_info, created_at, updated_at, deleted_at, requests_count, status
`

func (s *accessRequestStore) Create(ctx context.Context, newAccessRequest NewAccessRequest) (*types.AccessRequest, error) {
	// We don't allow adding a new request_access with an email address that has already been
	// verified by another user.
	exists, _, err := basestore.ScanFirstBool(s.Query(ctx, sqlf.Sprintf("SELECT TRUE WHERE EXISTS (SELECT FROM user_emails WHERE email = %s AND verified_at IS NOT NULL)", newAccessRequest.Email)))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errCannotCreateAccessRequest{errorCodeUserWithEmailExists}
	}

	// We don't allow adding a new request_access with an email address that has already been used
	res, err := s.ExecResult(ctx, sqlf.Sprintf("UPDATE access_requests SET requests_count = requests_count + 1 WHERE email = %s", newAccessRequest.Email))
	if err != nil {
		return nil, err
	}
	nrows, err := res.RowsAffected()
	if nrows > 0 {
		return nil, errCannotCreateAccessRequest{errorCodeAccessRequestWithEmailExists}
	}

	// Continue with creating the new access request.
	q := sqlf.Sprintf(
		accessRequestCreateQueryFmtStr,
		newAccessRequest.Name,
		newAccessRequest.Email,
		newAccessRequest.AdditionalInfo,
	)
	var accessRequest types.AccessRequest

	if err := s.QueryRow(ctx, q).Scan(
		&accessRequest.ID,
		&accessRequest.Name,
		&accessRequest.Email,
		&accessRequest.AdditionalInfo,
		&dbutil.NullTime{Time: &accessRequest.CreatedAt},
		&dbutil.NullTime{Time: &accessRequest.UpdatedAt},
		&dbutil.NullTime{Time: &accessRequest.DeletedAt},
		&accessRequest.RequestsCount,
		&accessRequest.Status,
	); err != nil {
		return nil, errors.Wrap(err, "scanning access_request")
	}

	return &accessRequest, nil

}

func (s *accessRequestStore) Delete(ctx context.Context, id int32) error {
	panic("implement me")
}

func (s *accessRequestStore) GetByID(ctx context.Context, id int32) (*types.User, error) {
	panic("implement me")
}

func (s *accessRequestStore) GetByEmail(ctx context.Context, email string) (*types.User, error) {
	panic("implement me")
}

func (s *accessRequestStore) HardDelete(ctx context.Context, id int32) error {
	panic("implement me")
}

func (s *accessRequestStore) Update(ctx context.Context, id int32, update UserUpdate) error {
	panic("implement me")
}

func (s *accessRequestStore) Count(ctx context.Context, opt *UsersListOptions) (int, error) {
	panic("implement me")
}

func (s *accessRequestStore) List(ctx context.Context, opt *UsersListOptions) ([]*types.User, error) {
	panic("implement me")
}
