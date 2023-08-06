package accessrequests

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/keegancsmith/sqlf"
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

type ListColumn string

const (
	ListID ListColumn = "id"
)

// ErrCannotCreate is the error that is returned when a request_access cannot be added to the DB due to a constraint.
type ErrCannotCreate struct {
	code string
}

func (err ErrCannotCreate) Error() string {
	return fmt.Sprintf("cannot create user: %v", err.code)
}

// ErrNotFound is the error that is returned when a request_access cannot be found in the DB.
type ErrNotFound struct {
	ID    int32
	Email string
}

func (e *ErrNotFound) Error() string {
	if e.Email != "" {
		return fmt.Sprintf("access_request with email %q not found", e.Email)
	}

	return fmt.Sprintf("access_request with ID %d not found", e.ID)
}

func (e *ErrNotFound) NotFound() bool {
	return true
}

// IsAccessRequestUserWithEmailExists reports whether err is an error indicating that the access request email was already taken by a signed in user.
func IsAccessRequestUserWithEmailExists(err error) bool {
	var e ErrCannotCreate
	return errors.As(err, &e) && e.code == errorCodeUserWithEmailExists
}

// IsAccessRequestWithEmailExists reports whether err is an error indicating that the access request was already created.
func IsAccessRequestWithEmailExists(err error) bool {
	var e ErrCannotCreate
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

type CreateQuery struct {
	AccessRequest *types.AccessRequest
}

type CreateResponse struct {
	AccessRequest *types.AccessRequest
}

func (q *CreateQuery) Execute(ctx context.Context, store *basestore.Store) (any, error) {
	var newAccessRequest *types.AccessRequest
	err := store.WithTransact(ctx, func(tx *basestore.Store) error {
		// We don't allow adding a new request_access with an email address that has already been
		// verified by another user.
		userExistsQuery := sqlf.Sprintf("SELECT TRUE FROM user_emails WHERE email = %s AND verified_at IS NOT NULL", q.AccessRequest.Email)
		exists, _, err := basestore.ScanFirstBool(tx.Query(ctx, userExistsQuery))
		if err != nil {
			return err
		}
		if exists {
			return ErrCannotCreate{errorCodeUserWithEmailExists}
		}

		// We don't allow adding a new request_access with an email address that has already been used
		accessRequestsExistsQuery := sqlf.Sprintf("SELECT TRUE FROM access_requests WHERE email = %s", q.AccessRequest.Email)
		exists, _, err = basestore.ScanFirstBool(tx.Query(ctx, accessRequestsExistsQuery))
		if err != nil {
			return err
		}
		if exists {
			return ErrCannotCreate{errorCodeAccessRequestWithEmailExists}
		}

		// Continue with creating the new access request.
		createQuery := sqlf.Sprintf(
			insertQuery,
			sqlf.Join(insertColumns, ","),
			q.AccessRequest.Name,
			q.AccessRequest.Email,
			q.AccessRequest.AdditionalInfo,
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
	return &CreateResponse{newAccessRequest}, err
}

func (c *ARClient) Create(ctx context.Context, accessRequest *types.AccessRequest) (*types.AccessRequest, error) {
	query := &CreateQuery{
		AccessRequest: accessRequest,
	}

	resp, err := c.dbclient.Execute(ctx, query)
	if err != nil {
		return nil, err
	}

	createResp, ok := resp.(*CreateResponse)
	if !ok {
		return nil, errors.New("Oh noes")
	}

	return createResp.AccessRequest, nil
}

type UpdateQuery struct {
	AccessRequest *types.AccessRequest
}

type UpdateResponse struct {
	AccessRequest *types.AccessRequest
}

func (q *UpdateQuery) Execute(ctx context.Context, store *basestore.Store) (any, error) {
	query := sqlf.Sprintf(updateQuery, q.AccessRequest.Status, *q.AccessRequest.DecisionByUserID, q.AccessRequest.ID, sqlf.Join(columns, ","))
	updated, err := scanAccessRequest(store.QueryRow(ctx, query))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &ErrNotFound{ID: q.AccessRequest.ID}
		}
		return nil, errors.Wrap(err, "scanning access_request")
	}

	return &UpdateResponse{updated}, nil
}

func (c *ARClient) Update(ctx context.Context, accessRequest *types.AccessRequest) (*types.AccessRequest, error) {
	query := &UpdateQuery{
		AccessRequest: accessRequest,
	}

	resp, err := c.dbclient.Execute(ctx, query)
	if err != nil {
		return nil, err
	}

	updateResp, ok := resp.(*UpdateResponse)
	if !ok {
		return nil, errors.New("Oh noes")
	}

	return updateResp.AccessRequest, nil
}

type GetByIDQuery struct {
	ID int32
}

type GetByIDResponse struct {
	AccessRequest *types.AccessRequest
}

func (c *ARClient) GetByID(ctx context.Context, id int32) (*types.AccessRequest, error) {
	query := &GetByIDQuery{
		ID: id,
	}

	resp, err := c.dbclient.Execute(ctx, query)
	if err != nil {
		return nil, err
	}

	getByIDResp, ok := resp.(*GetByIDResponse)
	if !ok {
		return nil, errors.New("Oh noes")
	}

	return getByIDResp.AccessRequest, nil
}

func (q *GetByIDQuery) Execute(ctx context.Context, store *basestore.Store) (any, error) {
	row := store.QueryRow(ctx, sqlf.Sprintf("SELECT %s FROM access_requests WHERE id = %s", sqlf.Join(columns, ","), q.ID))
	node, err := scanAccessRequest(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrNotFound{ID: q.ID}
		}
		return nil, err
	}

	return &GetByIDResponse{node}, nil
}

type GetByEmailQuery struct {
	Email string
}

type GetByEmailResponse struct {
	AccessRequest *types.AccessRequest
}

func (q *GetByEmailQuery) Execute(ctx context.Context, store *basestore.Store) (any, error) {
	row := store.QueryRow(ctx, sqlf.Sprintf("SELECT %s FROM access_requests WHERE email = %s", sqlf.Join(columns, ","), q.Email))
	node, err := scanAccessRequest(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrNotFound{Email: q.Email}
		}
		return nil, err
	}

	return &GetByEmailResponse{node}, nil
}

func (c *ARClient) GetByEmail(ctx context.Context, email string) (*types.AccessRequest, error) {
	query := &GetByEmailQuery{
		Email: email,
	}

	resp, err := c.dbclient.Execute(ctx, query)
	if err != nil {
		return nil, err
	}

	getByEmailResp, ok := resp.(*GetByEmailResponse)
	if !ok {
		return nil, errors.New("Oh noes")
	}

	return getByEmailResp.AccessRequest, nil
}

type CountQuery struct {
	FArgs *FilterArgs
}

type CountResponse struct {
	Count int
}

func (q *CountQuery) Execute(ctx context.Context, store *basestore.Store) (any, error) {
	query := sqlf.Sprintf("SELECT COUNT(*) FROM access_requests WHERE (%s)", sqlf.Join(q.FArgs.SQL(), ") AND ("))
	count, err := basestore.ScanInt(store.QueryRow(ctx, query))
	if err != nil {
		return nil, err
	}

	return &CountResponse{count}, nil
}

func (c *ARClient) Count(ctx context.Context, fArgs *FilterArgs) (int, error) {
	query := &CountQuery{
		FArgs: fArgs,
	}

	resp, err := c.dbclient.Execute(ctx, query)
	if err != nil {
		return 0, err
	}

	countResp, ok := resp.(*CountResponse)
	if !ok {
		return 0, errors.New("Oh noes")
	}

	return countResp.Count, nil
}

type ListQuery struct {
	FArgs *FilterArgs
	PArgs *database.PaginationArgs
}

type ListResponse struct {
	AccessRequests []*types.AccessRequest
}

func (q *ListQuery) Execute(ctx context.Context, store *basestore.Store) (any, error) {
	if q.FArgs == nil {
		q.FArgs = &FilterArgs{}
	}
	where := q.FArgs.SQL()
	if q.PArgs == nil {
		q.PArgs = &database.PaginationArgs{}
	}
	p := q.PArgs.SQL()

	if p.Where != nil {
		where = append(where, p.Where)
	}

	query := sqlf.Sprintf(listQuery, sqlf.Join(columns, ","), sqlf.Join(where, ") AND ("))
	query = p.AppendOrderToQuery(query)
	query = p.AppendLimitToQuery(query)

	nodes, err := scanAccessRequests(store.Query(ctx, query))
	if err != nil {
		return nil, err
	}

	return &ListResponse{nodes}, nil
}

func (c *ARClient) List(ctx context.Context, fArgs *FilterArgs, pArgs *database.PaginationArgs) ([]*types.AccessRequest, error) {
	query := &ListQuery{
		FArgs: fArgs,
		PArgs: pArgs,
	}

	resp, err := c.dbclient.Execute(ctx, query)
	if err != nil {
		return nil, err
	}

	listResp, ok := resp.(*ListResponse)
	if !ok {
		return nil, errors.New("Oh noes")
	}

	return listResp.AccessRequests, nil
}

func scanAccessRequest(sc dbutil.Scanner) (*types.AccessRequest, error) {
	var accessRequest types.AccessRequest
	if err := sc.Scan(&accessRequest.ID, &accessRequest.CreatedAt, &accessRequest.UpdatedAt, &accessRequest.Name, &accessRequest.Email, &accessRequest.Status, &accessRequest.AdditionalInfo, &accessRequest.DecisionByUserID); err != nil {
		return nil, err
	}

	return &accessRequest, nil
}

var scanAccessRequests = basestore.NewSliceScanner(scanAccessRequest)

type ARClient struct {
	dbclient database.DBClient
}

func NewARClient(dbclient database.DBClient) *ARClient {
	return &ARClient{dbclient}
}
