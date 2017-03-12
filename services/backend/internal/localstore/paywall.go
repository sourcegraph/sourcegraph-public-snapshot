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
	Trial   Plan = "trial"
	Paid    Plan = "paid"
	Blocked Plan = "blocked"

	orgTableName = "organization_payments"
)

type payments struct{}

func (payments) CreateTable() string {
	return `CREATE TABLE ` + orgTableName + ` (
			org_name text,
			plan text,
			trial_expiration timestamp,
			seats integer,
			PRIMARY KEY (org_name)
		);`
}

func (payments) DropTable() string {
	return "DROP TABLE IF EXISTS " + orgTableName + ";"
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
		return &Payment{Plan: Paid}, nil
	}
	var payment Payment
	err := appDBH(ctx).Db.QueryRow("SELECT plan, trial_expiration FROM "+orgTableName+" WHERE org_name = $1", repo.Owner).Scan(&payment.Plan, &payment.TrialExpiration)
	if err == sql.ErrNoRows {
		return &Payment{Plan: None}, nil
	}
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (p *payments) TrialExpirationDate(ctx context.Context, repo sourcegraph.Repo) (*time.Time, error) {
	if !repo.Private {
		return nil, nil
	}
	plan, err := p.paymentPlanForRepo(ctx, repo)
	if err != nil {
		return nil, err
	}
	if plan.Plan == Paid {
		return nil, nil
	}
	return plan.TrialExpiration, nil
}

func (p *payments) StartTrial(ctx context.Context, githubOrg string) error {
	_, err := appDBH(ctx).Db.Exec("INSERT INTO "+orgTableName+" (org_name, plan, trial_expiration) VALUES ($1, 'trial', LOCALTIMESTAMP + interval '14 days');", githubOrg)
	return err
}

func (p *payments) PaidForOrg(ctx context.Context, githubOrg string, seats uint64) error {
	_, err := appDBH(ctx).Db.Exec("UPDATE "+orgTableName+" SET plan = $1, seats = $2 WHERE org_name = $3;", Paid, seats, githubOrg)
	return err
}
