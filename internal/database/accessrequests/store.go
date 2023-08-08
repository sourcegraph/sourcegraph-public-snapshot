package accessrequests

import (
	"context"
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

var columns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
	sqlf.Sprintf("name"),
	sqlf.Sprintf("email"),
	sqlf.Sprintf("status"),
	sqlf.Sprintf("additional_info"),
	sqlf.Sprintf("decision_by_user_id"),
}

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

type Store struct {
	dbStore database.DBStore
}

func NewStore(dbStore database.DBStore) *Store {
	return &Store{dbStore}
}

func (c *Store) Create(ctx context.Context, accessRequest *types.AccessRequest) (*types.AccessRequest, error) {
	command := &Create{
		AccessRequest: accessRequest,
	}

	if err := c.dbStore.ExecuteCommand(ctx, command); err != nil {
		return nil, err
	}

	return command.Response, nil
}

func (c *Store) Update(ctx context.Context, accessRequest *types.AccessRequest) (*types.AccessRequest, error) {
	command := &Update{
		AccessRequest: accessRequest,
	}

	if err := c.dbStore.ExecuteCommand(ctx, command); err != nil {
		return nil, err
	}

	return command.Response, nil
}

func (c *Store) GetByID(ctx context.Context, id int32) (*types.AccessRequest, error) {
	command := &GetByID{
		ID: id,
	}

	if err := c.dbStore.ExecuteCommand(ctx, command); err != nil {
		return nil, err
	}

	return command.Response, nil
}

func (c *Store) Count(ctx context.Context, fArgs *FilterArgs) (int, error) {
	command := &Count{
		FArgs: fArgs,
	}

	if err := c.dbStore.ExecuteCommand(ctx, command); err != nil {
		return 0, err
	}

	return command.Response, nil
}

func (c *Store) List(ctx context.Context, fArgs *FilterArgs, pArgs *database.PaginationArgs) ([]*types.AccessRequest, error) {
	command := &List{
		FArgs: fArgs,
		PArgs: pArgs,
	}

	if err := c.dbStore.ExecuteCommand(ctx, command); err != nil {
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
