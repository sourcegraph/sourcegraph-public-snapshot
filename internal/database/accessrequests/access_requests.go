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

// IsUserWithEmailExists reports whether err is an error indicating that the access request email was already taken by a signed in user.
func IsUserWithEmailExists(err error) bool {
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

type Client struct {
	dbclient database.DBClient
}

func NewClient(dbclient database.DBClient) *Client {
	return &Client{dbclient}
}

type Create struct {
	AccessRequest *types.AccessRequest

	Response *types.AccessRequest
}

func (c *Create) Execute(ctx context.Context, store *basestore.Store) error {
	err := store.WithTransact(ctx, func(tx *basestore.Store) error {
		// We don't allow adding a new request_access with an email address that has already been
		// verified by another user.
		userExistsQuery := sqlf.Sprintf("SELECT TRUE FROM user_emails WHERE email = %s AND verified_at IS NOT NULL", c.AccessRequest.Email)
		exists, _, err := basestore.ScanFirstBool(tx.Query(ctx, userExistsQuery))
		if err != nil {
			return err
		}
		if exists {
			return ErrCannotCreate{errorCodeUserWithEmailExists}
		}

		// We don't allow adding a new request_access with an email address that has already been used
		accessRequestsExistsQuery := sqlf.Sprintf("SELECT TRUE FROM access_requests WHERE email = %s", c.AccessRequest.Email)
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
			c.AccessRequest.Name,
			c.AccessRequest.Email,
			c.AccessRequest.AdditionalInfo,
			types.AccessRequestStatusPending,
			sqlf.Join(columns, ","),
		)
		data, err := scanAccessRequest(tx.QueryRow(ctx, createQuery))
		c.Response = data
		if err != nil {
			return errors.Wrap(err, "scanning access_request")
		}

		return nil
	})
	return err
}

func (c *Client) Create(ctx context.Context, accessRequest *types.AccessRequest) (*types.AccessRequest, error) {
	command := &Create{
		AccessRequest: accessRequest,
	}

	if err := c.dbclient.ExecuteCommand(ctx, command); err != nil {
		return nil, err
	}

	return command.Response, nil
}

type Update struct {
	AccessRequest *types.AccessRequest

	Response *types.AccessRequest
}

func (c *Update) Execute(ctx context.Context, store *basestore.Store) error {
	query := sqlf.Sprintf(updateQuery, c.AccessRequest.Status, *c.AccessRequest.DecisionByUserID, c.AccessRequest.ID, sqlf.Join(columns, ","))
	updated, err := scanAccessRequest(store.QueryRow(ctx, query))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &ErrNotFound{ID: c.AccessRequest.ID}
		}
		return errors.Wrap(err, "scanning access_request")
	}

	c.Response = updated
	return nil
}

func (c *Client) Update(ctx context.Context, accessRequest *types.AccessRequest) (*types.AccessRequest, error) {
	command := &Update{
		AccessRequest: accessRequest,
	}

	if err := c.dbclient.ExecuteCommand(ctx, command); err != nil {
		return nil, err
	}

	return command.Response, nil
}

type GetByID struct {
	ID int32

	Response *types.AccessRequest
}

func (c *GetByID) Execute(ctx context.Context, store *basestore.Store) error {
	row := store.QueryRow(ctx, sqlf.Sprintf("SELECT %s FROM access_requests WHERE id = %s", sqlf.Join(columns, ","), c.ID))
	node, err := scanAccessRequest(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return &ErrNotFound{ID: c.ID}
		}
		return err
	}

	c.Response = node
	return nil
}

func (c *Client) GetByID(ctx context.Context, id int32) (*types.AccessRequest, error) {
	command := &GetByID{
		ID: id,
	}

	if err := c.dbclient.ExecuteCommand(ctx, command); err != nil {
		return nil, err
	}

	return command.Response, nil
}

type GetByEmail struct {
	Email string

	Response *types.AccessRequest
}

func (c *GetByEmail) Execute(ctx context.Context, store *basestore.Store) error {
	row := store.QueryRow(ctx, sqlf.Sprintf("SELECT %s FROM access_requests WHERE email = %s", sqlf.Join(columns, ","), c.Email))
	node, err := scanAccessRequest(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return &ErrNotFound{Email: c.Email}
		}
		return err
	}

	c.Response = node
	return nil
}

func (c *Client) GetByEmail(ctx context.Context, email string) (*types.AccessRequest, error) {
	command := &GetByEmail{
		Email: email,
	}

	if err := c.dbclient.ExecuteCommand(ctx, command); err != nil {
		return nil, err
	}

	return command.Response, nil
}

type Count struct {
	FArgs *FilterArgs

	Response int
}

func (c *Count) Execute(ctx context.Context, store *basestore.Store) error {
	query := sqlf.Sprintf("SELECT COUNT(*) FROM access_requests WHERE (%s)", sqlf.Join(c.FArgs.SQL(), ") AND ("))
	count, err := basestore.ScanInt(store.QueryRow(ctx, query))
	if err != nil {
		return err
	}

	c.Response = count
	return nil
}

func (c *Client) Count(ctx context.Context, fArgs *FilterArgs) (int, error) {
	command := &Count{
		FArgs: fArgs,
	}

	if err := c.dbclient.ExecuteCommand(ctx, command); err != nil {
		return 0, err
	}

	return command.Response, nil
}

type List struct {
	FArgs *FilterArgs
	PArgs *database.PaginationArgs

	Response []*types.AccessRequest
}

func (c *List) Execute(ctx context.Context, store *basestore.Store) error {
	if c.FArgs == nil {
		c.FArgs = &FilterArgs{}
	}
	where := c.FArgs.SQL()
	if c.PArgs == nil {
		c.PArgs = &database.PaginationArgs{}
	}
	p := c.PArgs.SQL()

	if p.Where != nil {
		where = append(where, p.Where)
	}

	query := sqlf.Sprintf(listQuery, sqlf.Join(columns, ","), sqlf.Join(where, ") AND ("))
	query = p.AppendOrderToQuery(query)
	query = p.AppendLimitToQuery(query)

	nodes, err := scanAccessRequests(store.Query(ctx, query))
	if err != nil {
		return err
	}

	c.Response = nodes
	return nil
}

func (c *Client) List(ctx context.Context, fArgs *FilterArgs, pArgs *database.PaginationArgs) ([]*types.AccessRequest, error) {
	command := &List{
		FArgs: fArgs,
		PArgs: pArgs,
	}

	if err := c.dbclient.ExecuteCommand(ctx, command); err != nil {
		return nil, err
	}

	return command.Response, nil
}

func scanAccessRequest(sc dbutil.Scanner) (*types.AccessRequest, error) {
	var accessRequest types.AccessRequest
	if err := sc.Scan(&accessRequest.ID, &accessRequest.CreatedAt, &accessRequest.UpdatedAt, &accessRequest.Name, &accessRequest.Email, &accessRequest.Status, &accessRequest.AdditionalInfo, &accessRequest.DecisionByUserID); err != nil {
		return nil, err
	}

	return &accessRequest, nil
}

var scanAccessRequests = basestore.NewSliceScanner(scanAccessRequest)
