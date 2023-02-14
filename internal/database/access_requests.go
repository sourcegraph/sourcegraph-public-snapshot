package database

import (
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
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

type AccessRequestsFilterOptions struct {
	Status *types.AccessRequestStatus
}

func (o AccessRequestsFilterOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{
		sqlf.Sprintf("deleted_at IS NULL"),
	}
	if o.Status != nil {
		conds = append(conds, sqlf.Sprintf("status = %v", *o.Status))
	}
	return conds
}

type AccessRequestsListOptions struct {
	OrderBy    *string
	Descending *bool
	Limit      *int32
	Offset     *int32
}

func (o AccessRequestsListOptions) sqlOrderBy() (*sqlf.Query, error) {
	orderDirection := "ASC"
	if o.Descending != nil && *o.Descending {
		orderDirection = "DESC"
	}
	orderBy := sqlf.Sprintf("id " + orderDirection)
	if o.OrderBy != nil {
		newOrderColumn, err := toAccessRequestsField(*o.OrderBy)
		orderBy = sqlf.Sprintf(newOrderColumn + " " + orderDirection)
		if err != nil {
			return nil, err
		}
	}

	return orderBy, nil
}

func (o AccessRequestsListOptions) sqlLimit() *sqlf.Query {
	limit := int32(100)
	if o.Limit != nil {
		limit = *o.Limit
	}

	offset := int32(0)
	if o.Offset != nil {
		offset = *o.Offset
	}

	return sqlf.Sprintf(`%s OFFSET %s`, limit, offset)
}

type AccessRequestsFilterAndListOptions struct {
	AccessRequestsListOptions
	AccessRequestsFilterOptions
}

func toAccessRequestsField(orderBy string) (string, error) {
	switch orderBy {
	case "NAME":
		return "name", nil
	case "EMAIL":
		return "email", nil
	case "CREATED_AT":
		return "created_at", nil
	default:
		return "", errors.New("invalid orderBy")
	}
}

// AccessRequestStore provides access to the `access_requests` table.
//
// For a detailed overview of the schema, see schema.md.
type AccessRequestStore interface {
	basestore.ShareableStore
	Count(context.Context, AccessRequestsFilterOptions) (int, error)
	Create(context.Context, NewAccessRequest) (*types.AccessRequest, error)
	Delete(context.Context, int32) error
	GetByID(context.Context, int32) (*types.AccessRequest, error)
	GetByEmail(context.Context, string) (*types.AccessRequest, error)
	HardDelete(context.Context, int32) error
	List(context.Context, AccessRequestsFilterAndListOptions) (_ []*types.AccessRequest, err error)
	Update(context.Context, int32, UserUpdate) error
}

type accessRequestStore struct {
	*basestore.Store
	logger log.Logger
}

type NewAccessRequest struct {
	Name           string
	Email          string
	AdditionalInfo string
}

// AccessRequestsWith instantiates and returns a new accessRequestStore using the other store handle.
func AccessRequestsWith(other basestore.ShareableStore, logger log.Logger) AccessRequestStore {
	return &accessRequestStore{Store: basestore.NewWithHandle(other.Handle()), logger: logger}
}

const (
	accessRequestInsertQuery = `
		INSERT INTO
			access_requests (name, email, additional_info)
		VALUES ( %s, %s, %s )
		RETURNING id, created_at, updated_at, deleted_at, name, email, status, additional_info
		`
	accessRequestListQuery = `
		SELECT
			id, created_at, updated_at, deleted_at, name, email, status, additional_info
		FROM
			access_requests
		WHERE (%s)
		ORDER BY %s
		LIMIT %s
		`
)

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
	exists, _, err = basestore.ScanFirstBool(s.Query(ctx, sqlf.Sprintf("SELECT TRUE WHERE EXISTS (SELECT FROM access_requests WHERE email = %s)", newAccessRequest.Email)))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errCannotCreateAccessRequest{errorCodeAccessRequestWithEmailExists}
	}

	// Continue with creating the new access request.
	q := sqlf.Sprintf(
		accessRequestInsertQuery,
		newAccessRequest.Name,
		newAccessRequest.Email,
		newAccessRequest.AdditionalInfo,
	)
	var data types.AccessRequest

	if err := s.QueryRow(ctx, q).Scan(&data.ID, &data.CreatedAt, &data.UpdatedAt, &data.DeletedAt, &data.Name, &data.Email, &data.Status, &data.AdditionalInfo); err != nil {
		return nil, errors.Wrap(err, "scanning access_request")
	}

	return &data, nil

}

func (s *accessRequestStore) Delete(ctx context.Context, id int32) error {
	panic("implement me")
}

func (s *accessRequestStore) GetByID(ctx context.Context, id int32) (*types.AccessRequest, error) {
	panic("implement me")
}

func (s *accessRequestStore) GetByEmail(ctx context.Context, email string) (*types.AccessRequest, error) {
	panic("implement me")
}

func (s *accessRequestStore) HardDelete(ctx context.Context, id int32) error {
	panic("implement me")
}

func (s *accessRequestStore) Update(ctx context.Context, id int32, update UserUpdate) error {
	panic("implement me")
}

func (s *accessRequestStore) Count(ctx context.Context, opt AccessRequestsFilterOptions) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM access_requests WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := s.QueryRow(ctx, q).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *accessRequestStore) List(ctx context.Context, opt AccessRequestsFilterAndListOptions) ([]*types.AccessRequest, error) {
	orderBy, err := opt.sqlOrderBy()
	if err != nil {
		return nil, err
	}

	fmt.Println(sqlf.Join(opt.sqlConditions(), ") AND ("))
	query := sqlf.Sprintf(accessRequestListQuery, sqlf.Join(opt.sqlConditions(), ") AND ("), orderBy, opt.sqlLimit())

	rows, err := s.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	nodes := make([]*types.AccessRequest, 0)
	for rows.Next() {
		var node types.AccessRequest

		if err := rows.Scan(&node.ID, &node.CreatedAt, &node.UpdatedAt, &node.DeletedAt, &node.Name, &node.Email, &node.Status, &node.AdditionalInfo); err != nil {
			return nil, err
		}

		nodes = append(nodes, &node)
	}

	return nodes, nil
}
