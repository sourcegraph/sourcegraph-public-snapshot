package database

import (
	"context"
	"database/sql"
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

// ErrCannotCreateAccessRequest is the error that is returned when a request_access cannot be added to the DB due to a constraint.
type ErrCannotCreateAccessRequest struct {
	code string
}

func (err ErrCannotCreateAccessRequest) Error() string {
	return fmt.Sprintf("cannot create user: %v", err.code)
}

// ErrAccessRequestNotFound is the error that is returned when a request_access cannot be found in the DB.
type ErrAccessRequestNotFound struct {
	ID    int32
	Email string
}

func (e *ErrAccessRequestNotFound) Error() string {
	if e.Email != "" {
		return fmt.Sprintf("access_request with email %q not found", e.Email)
	}

	return fmt.Sprintf("access_request with ID %d not found", e.ID)
}

func (e *ErrAccessRequestNotFound) NotFound() bool {
	return true
}

// IsAccessRequestUserWithEmailExists reports whether err is an error indicating that the access request email was already taken by a signed in user.
func IsAccessRequestUserWithEmailExists(err error) bool {
	var e ErrCannotCreateAccessRequest
	return errors.As(err, &e) && e.code == errorCodeUserWithEmailExists
}

// IsAccessRequestWithEmailExists reports whether err is an error indicating that the access request was already created.
func IsAccessRequestWithEmailExists(err error) bool {
	var e ErrCannotCreateAccessRequest
	return errors.As(err, &e) && e.code == errorCodeAccessRequestWithEmailExists
}

type AccessRequestsFilterOptions struct {
	Status *types.AccessRequestStatus
}

func (o *AccessRequestsFilterOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o != nil && o.Status != nil {
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

func (o *AccessRequestsListOptions) sqlOrderBy() (*sqlf.Query, error) {
	orderDirection := "ASC"
	if o != nil && o.Descending != nil && *o.Descending {
		orderDirection = "DESC"
	}
	orderBy := sqlf.Sprintf("id " + orderDirection)
	if o != nil && o.OrderBy != nil {
		newOrderColumn, err := toAccessRequestsField(*o.OrderBy)
		orderBy = sqlf.Sprintf(newOrderColumn + " " + orderDirection)
		if err != nil {
			return nil, err
		}
	}

	return orderBy, nil
}

func (o *AccessRequestsListOptions) sqlLimit() *sqlf.Query {
	limit := int32(100)
	if o != nil && o.Limit != nil {
		limit = *o.Limit
	}

	offset := int32(0)
	if o != nil && o.Offset != nil {
		offset = *o.Offset
	}

	return sqlf.Sprintf(`%s OFFSET %s`, limit, offset)
}

type AccessRequestsFilterAndListOptions struct {
	*AccessRequestsListOptions
	*AccessRequestsFilterOptions
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
	Create(context.Context, *types.AccessRequest) (*types.AccessRequest, error)
	Update(context.Context, *types.AccessRequest) (*types.AccessRequest, error)
	GetByID(context.Context, int32) (*types.AccessRequest, error)
	GetByEmail(context.Context, string) (*types.AccessRequest, error)
	Count(context.Context, *AccessRequestsFilterOptions) (int, error)
	List(context.Context, *AccessRequestsFilterAndListOptions) (_ []*types.AccessRequest, err error)
}

type accessRequestStore struct {
	*basestore.Store
	logger log.Logger
}

// AccessRequestsWith instantiates and returns a new accessRequestStore using the other store handle.
func AccessRequestsWith(other basestore.ShareableStore, logger log.Logger) AccessRequestStore {
	return &accessRequestStore{Store: basestore.NewWithHandle(other.Handle()), logger: logger}
}

const (
	accessRequestInsertQuery = `
		INSERT INTO access_requests (%s)
		VALUES ( %s, %s, %s )
		RETURNING %s`
	accessRequestListQuery = `
		SELECT %s
		FROM access_requests
		WHERE (%s)
		ORDER BY %s
		LIMIT %s`
	accessRequestUpdateQuery = `
		UPDATE access_requests
		SET status = %s
		WHERE id = %s
		RETURNING %s`
)

var (
	accessRequestColumns = []*sqlf.Query{
		sqlf.Sprintf("id"),
		sqlf.Sprintf("created_at"),
		sqlf.Sprintf("updated_at"),
		sqlf.Sprintf("name"),
		sqlf.Sprintf("email"),
		sqlf.Sprintf("status"),
		sqlf.Sprintf("additional_info"),
	}
	accessRequestInsertColumns = []*sqlf.Query{
		sqlf.Sprintf("name"),
		sqlf.Sprintf("email"),
		sqlf.Sprintf("additional_info"),
	}
)

func (s *accessRequestStore) Create(ctx context.Context, accessRequest *types.AccessRequest) (*types.AccessRequest, error) {
	// We don't allow adding a new request_access with an email address that has already been
	// verified by another user.
	exists, _, err := basestore.ScanFirstBool(s.Query(ctx, sqlf.Sprintf("SELECT TRUE FROM user_emails WHERE email = %s AND verified_at IS NOT NULL", accessRequest.Email)))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCannotCreateAccessRequest{errorCodeUserWithEmailExists}
	}

	// We don't allow adding a new request_access with an email address that has already been used
	exists, _, err = basestore.ScanFirstBool(s.Query(ctx, sqlf.Sprintf("SELECT TRUE FROM access_requests WHERE email = %s", accessRequest.Email)))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCannotCreateAccessRequest{errorCodeAccessRequestWithEmailExists}
	}

	// Continue with creating the new access request.
	q := sqlf.Sprintf(
		accessRequestInsertQuery,
		sqlf.Join(accessRequestInsertColumns, ","),
		accessRequest.Name,
		accessRequest.Email,
		accessRequest.AdditionalInfo,
		sqlf.Join(accessRequestColumns, ","),
	)

	data, err := scanAccessRequest(s.QueryRow(ctx, q))
	if err != nil {
		return nil, errors.Wrap(err, "scanning access_request")
	}

	return data, nil
}

func (s *accessRequestStore) GetByID(ctx context.Context, id int32) (*types.AccessRequest, error) {
	row := s.QueryRow(ctx, sqlf.Sprintf("SELECT %s FROM access_requests WHERE id = %s", sqlf.Join(accessRequestColumns, ","), id))
	node, err := scanAccessRequest(row)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrAccessRequestNotFound{ID: id}
		}
		return nil, err
	}

	return node, nil
}

func (s *accessRequestStore) GetByEmail(ctx context.Context, email string) (*types.AccessRequest, error) {
	row := s.QueryRow(ctx, sqlf.Sprintf("SELECT %s FROM access_requests WHERE email = %s", sqlf.Join(accessRequestColumns, ","), email))
	node, err := scanAccessRequest(row)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrAccessRequestNotFound{Email: email}
		}
		return nil, err
	}

	return node, nil
}

func (s *accessRequestStore) Update(ctx context.Context, accessRequest *types.AccessRequest) (*types.AccessRequest, error) {
	q := sqlf.Sprintf(accessRequestUpdateQuery, accessRequest.Status, accessRequest.ID, sqlf.Join(accessRequestColumns, ","))
	updated, err := scanAccessRequest(s.QueryRow(ctx, q))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &ErrAccessRequestNotFound{ID: accessRequest.ID}
		}
		return nil, errors.Wrap(err, "scanning access_request")
	}

	return updated, nil
}

func (s *accessRequestStore) Count(ctx context.Context, opt *AccessRequestsFilterOptions) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM access_requests WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	return basestore.ScanInt(s.QueryRow(ctx, q))
}

func (s *accessRequestStore) List(ctx context.Context, opt *AccessRequestsFilterAndListOptions) ([]*types.AccessRequest, error) {
	orderBy, err := opt.sqlOrderBy()
	if err != nil {
		return nil, err
	}

	query := sqlf.Sprintf(accessRequestListQuery,
		sqlf.Join(accessRequestColumns, ","),
		sqlf.Join(opt.sqlConditions(), ") AND ("), orderBy, opt.sqlLimit())

	nodes, err := scanAccessRequests(s.Query(ctx, query))
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func scanAccessRequest(sc dbutil.Scanner) (*types.AccessRequest, error) {
	var accessRequest types.AccessRequest
	if err := sc.Scan(&accessRequest.ID, &accessRequest.CreatedAt, &accessRequest.UpdatedAt, &accessRequest.Name, &accessRequest.Email, &accessRequest.Status, &accessRequest.AdditionalInfo); err != nil {
		return nil, err
	}

	return &accessRequest, nil
}

var scanAccessRequests = basestore.NewSliceScanner(scanAccessRequest)
