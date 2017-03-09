package localstore

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
)

type Plan string

const (
	None    Plan = "none"
	Blocked Plan = "blocked"

	personalTableName = "personal_payments"
	orgTableName      = "organization_payments"
)

type payments struct{}

func (payments) CreateTable() string {
	return `CREATE TABLE ` + personalTableName + ` (
			user_id text,
			plan text,
			trial_expiration timestamp,
			PRIMARY KEY (user_id)
		);
		CREATE TABLE ` + orgTableName + ` (
			org_name text,
			plan text,
			trial_expiration timestamp,
			PRIMARY KEY (org_name)
		);`
}

func (payments) DropTable() string {
	return "DROP TABLE IF EXISTS " + personalTableName + ";" +
		"DROP TABLE IF EXISTS " + orgTableName + ";"
}

type Payment struct {
	Plan            Plan       `db:"plan"`
	TrialExpiration *time.Time `db:"trial_expiration"`
}

func (p *payments) paymentPlanForRepo(ctx context.Context, repo sourcegraph.Repo) (*Payment, error) {
	actor := auth.ActorFromContext(ctx)
	if actor.Login == "" {
		return nil, errors.New("user must have a login to access private repos")
	}
	if actor.Login == repo.Owner {
		return p.getPersonalPlan(ctx)
	}
	return p.getOrgPlan(ctx, repo.Owner)
}

func (p *payments) getPersonalPlan(ctx context.Context) (*Payment, error) {
	var payment *Payment
	actor := auth.ActorFromContext(ctx)
	err := appDBH(ctx).Db.QueryRow("SELECT * FROM "+personalTableName+" WHERE user_id = $1", actor.UID).Scan(payment)
	if err == sql.ErrNoRows {
		return &Payment{Plan: None}, nil
	}
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (p *payments) getOrgPlan(ctx context.Context, owner string) (*Payment, error) {
	var payment Payment
	err := appDBH(ctx).Db.QueryRow("SELECT plan, trial_expiration FROM "+orgTableName+" WHERE org_name = $1", owner).Scan(&payment.Plan, &payment.TrialExpiration)
	if err == sql.ErrNoRows {
		return &Payment{Plan: None}, nil
	}
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// CheckPaywallForRepo returns an error if the user is not on a payment plan that
// allows them to see this kind of repository. It is not critical to call in
// all paths, just primary graphQL paths, i.e. it is OK if an unpaid user sees
// references from a private repo, but they shouldn't be able to browse it.
func (p *payments) CheckPaywallForRepo(ctx context.Context, repo sourcegraph.Repo) error {
	if !repo.Private {
		return nil
	}

	payment, err := p.paymentPlanForRepo(ctx, repo)
	if err != nil {
		return err
	}

	if payment == nil {
		payment = &Payment{Plan: None}
	}

	if payment.Plan == Blocked {
		return ErrBlocked{}
	}

	expiration, err := p.TrialExpirationDate(ctx, repo)
	if err != nil {
		return err
	}
	if expiration == nil {
		return nil
	}
	if expiration.Before(time.Now()) {
		return ErrTrialExpired{}
	}
	return nil
}

type ErrBlocked struct{}

func (ErrBlocked) Error() string {
	return "account blocked: this account has not paid for private repos"
}

type ErrTrialExpired struct{}

func (ErrTrialExpired) Error() string {
	return "trial expired: this account has not paid for private repos"
}

func (p *payments) TrialExpirationDate(ctx context.Context, repo sourcegraph.Repo) (*time.Time, error) {
	if !repo.Private {
		return nil, nil
	}
	plan, err := p.paymentPlanForRepo(ctx, repo)
	if err != nil {
		return nil, err
	}
	return plan.TrialExpiration, nil
}
