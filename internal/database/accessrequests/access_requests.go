package accessrequests

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
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

type FilterArgs struct {
	Status *types.AccessRequestStatus
}

func (o *FilterArgs) SQL() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o != nil && o.Status != nil {
		conds = append(conds, sqlf.Sprintf("status = %v", *o.Status))
	}
	return conds
}

// Store provides access to the `access_requests` table.
//
// For a detailed overview of the schema, see schema.md.
type Store interface {
	basestore.ShareableStore
	Create(context.Context, *types.AccessRequest) (*types.AccessRequest, error)
	Update(context.Context, *types.AccessRequest) (*types.AccessRequest, error)
	GetByID(context.Context, int32) (*types.AccessRequest, error)
	GetByEmail(context.Context, string) (*types.AccessRequest, error)
	Count(context.Context, *FilterArgs) (int, error)
	List(context.Context, *FilterArgs, *database.PaginationArgs) (_ []*types.AccessRequest, err error)
	WithTransact(context.Context, func(Store) error) error
	Done(error) error
}

type store struct {
	*basestore.Store
	logger log.Logger
}

// With instantiates and returns a new accessRequestStore using the other store handle.
func With(other basestore.ShareableStore, logger log.Logger) Store {
	if mockStore != nil {
		return mockStore
	}
	return &store{Store: basestore.NewWithHandle(other.Handle()), logger: logger}
}

const (
	insertQuery = `
		INSERT INTO access_requests (%s)
		VALUES ( %s, %s, %s, %s )
		RETURNING %s`
	listQuery = `
		SELECT %s
		FROM access_requests
		WHERE (%s)`
	updateQuery = `
		UPDATE access_requests
		SET status = %s, updated_at = NOW(), decision_by_user_id = %s
		WHERE id = %s
		RETURNING %s`
)

type ListColumn string

const (
	ListID ListColumn = "id"
)

var (
	columns = []*sqlf.Query{
		sqlf.Sprintf("id"),
		sqlf.Sprintf("created_at"),
		sqlf.Sprintf("updated_at"),
		sqlf.Sprintf("name"),
		sqlf.Sprintf("email"),
		sqlf.Sprintf("status"),
		sqlf.Sprintf("additional_info"),
		sqlf.Sprintf("decision_by_user_id"),
	}
	insertColumns = []*sqlf.Query{
		sqlf.Sprintf("name"),
		sqlf.Sprintf("email"),
		sqlf.Sprintf("additional_info"),
		sqlf.Sprintf("status"),
	}
)

func (s *store) Create(ctx context.Context, accessRequest *types.AccessRequest) (*types.AccessRequest, error) {
	var newAccessRequest *types.AccessRequest
	err := s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		// We don't allow adding a new request_access with an email address that has already been
		// verified by another user.
		userExistsQuery := sqlf.Sprintf("SELECT TRUE FROM user_emails WHERE email = %s AND verified_at IS NOT NULL", accessRequest.Email)
		exists, _, err := basestore.ScanFirstBool(tx.Query(ctx, userExistsQuery))
		if err != nil {
			return err
		}
		if exists {
			return ErrCannotCreateAccessRequest{errorCodeUserWithEmailExists}
		}

		// We don't allow adding a new request_access with an email address that has already been used
		accessRequestsExistsQuery := sqlf.Sprintf("SELECT TRUE FROM access_requests WHERE email = %s", accessRequest.Email)
		exists, _, err = basestore.ScanFirstBool(tx.Query(ctx, accessRequestsExistsQuery))
		if err != nil {
			return err
		}
		if exists {
			return ErrCannotCreateAccessRequest{errorCodeAccessRequestWithEmailExists}
		}

		// Continue with creating the new access request.
		createQuery := sqlf.Sprintf(
			insertQuery,
			sqlf.Join(insertColumns, ","),
			accessRequest.Name,
			accessRequest.Email,
			accessRequest.AdditionalInfo,
			types.AccessRequestStatusPending,
			sqlf.Join(columns, ","),
		)
		data, err := scanAccessRequest(tx.QueryRow(ctx, createQuery))
		newAccessRequest = data
		if err != nil {
			return errors.Wrap(err, "scanning access_request")
		}

		return nil
	})
	return newAccessRequest, err
}

func (s *store) GetByID(ctx context.Context, id int32) (*types.AccessRequest, error) {
	row := s.QueryRow(ctx, sqlf.Sprintf("SELECT %s FROM access_requests WHERE id = %s", sqlf.Join(columns, ","), id))
	node, err := scanAccessRequest(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrAccessRequestNotFound{ID: id}
		}
		return nil, err
	}

	return node, nil
}

func (s *store) GetByEmail(ctx context.Context, email string) (*types.AccessRequest, error) {
	row := s.QueryRow(ctx, sqlf.Sprintf("SELECT %s FROM access_requests WHERE email = %s", sqlf.Join(columns, ","), email))
	node, err := scanAccessRequest(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrAccessRequestNotFound{Email: email}
		}
		return nil, err
	}

	return node, nil
}

func (s *store) Update(ctx context.Context, accessRequest *types.AccessRequest) (*types.AccessRequest, error) {
	q := sqlf.Sprintf(updateQuery, accessRequest.Status, *accessRequest.DecisionByUserID, accessRequest.ID, sqlf.Join(columns, ","))
	updated, err := scanAccessRequest(s.QueryRow(ctx, q))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &ErrAccessRequestNotFound{ID: accessRequest.ID}
		}
		return nil, errors.Wrap(err, "scanning access_request")
	}

	return updated, nil
}

func (s *store) Count(ctx context.Context, fArgs *FilterArgs) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM access_requests WHERE (%s)", sqlf.Join(fArgs.SQL(), ") AND ("))
	return basestore.ScanInt(s.QueryRow(ctx, q))
}

func (s *store) List(ctx context.Context, fArgs *FilterArgs, pArgs *database.PaginationArgs) ([]*types.AccessRequest, error) {
	if fArgs == nil {
		fArgs = &FilterArgs{}
	}
	where := fArgs.SQL()
	if pArgs == nil {
		pArgs = &database.PaginationArgs{}
	}
	p := pArgs.SQL()

	if p.Where != nil {
		where = append(where, p.Where)
	}

	q := sqlf.Sprintf(listQuery, sqlf.Join(columns, ","), sqlf.Join(where, ") AND ("))
	q = p.AppendOrderToQuery(q)
	q = p.AppendLimitToQuery(q)

	nodes, err := scanAccessRequests(s.Query(ctx, q))
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func (s *store) WithTransact(ctx context.Context, f func(tx Store) error) error {
	return s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(&store{
			logger: s.logger,
			Store:  tx,
		})
	})
}

func scanAccessRequest(sc dbutil.Scanner) (*types.AccessRequest, error) {
	var accessRequest types.AccessRequest
	if err := sc.Scan(&accessRequest.ID, &accessRequest.CreatedAt, &accessRequest.UpdatedAt, &accessRequest.Name, &accessRequest.Email, &accessRequest.Status, &accessRequest.AdditionalInfo, &accessRequest.DecisionByUserID); err != nil {
		return nil, err
	}

	return &accessRequest, nil
}

var scanAccessRequests = basestore.NewSliceScanner(scanAccessRequest)
