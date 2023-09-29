package codehost_scenario

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v53/github"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Repo struct {
	s    *GithubScenario
	team *Team
	org  *Org
	name string
}

func (r *Repo) Get(ctx context.Context) (*github.Repository, error) {
	if r.s.IsApplied() {
		return r.get(ctx)
	}
	panic("cannot retrieve repo before scenario is applied")
}

func (r *Repo) get(ctx context.Context) (*github.Repository, error) {
	return r.s.client.GetRepo(ctx, r.org.name, r.name)
}

func (r *Repo) AddTeam(team *Team) {
	r.team = team
	action := &action{
		name: fmt.Sprintf("repo:team:%s:membership:%s", team.name, r.name),
		apply: func(ctx context.Context) error {
			org, err := r.org.get(ctx)
			if err != nil {
				return err
			}

			repo, err := r.get(ctx)
			if err != nil {
				return err
			}

			team, err := r.team.get(ctx)
			if err != nil {
				return err
			}

			err = r.s.client.UpdateTeamRepoPermissions(ctx, org, team, repo)
			if err != nil {
				return err
			}
			return nil
		},
		teardown: nil,
	}

	r.s.append(action)
}

func (r *Repo) SetPermissions(private bool) {
	permissionKey := "private"
	if !private {
		permissionKey = "public"
	}
	action := &action{
		name: fmt.Sprintf("repo:permissions:%s:%s", r.name, permissionKey),
		apply: func(ctx context.Context) error {
			repo, err := r.get(ctx)
			if err != nil {
				return err
			}
			repo.Private = &private

			org, err := r.org.get(ctx)
			if err != nil {
				return err
			}

			_, err = r.s.client.UpdateRepo(ctx, org, repo)
			if err != nil {
				return err
			}
			return err
		},
	}

	r.s.append(action)
}

func (r *Repo) WaitTillExists() {
	action := &action{
		name: fmt.Sprintf("repo:exists:%s", r.name),
		apply: func(ctx context.Context) error {
			var err error
			for i := 0; i < 5; i++ {
				time.Sleep(1 * time.Second)
				_, err = r.get(ctx)
				if err == nil {
					return nil
				}
			}
			return errors.Newf("repo %q did not exist after waiting: %v", r.name, err)
		},
	}

	r.s.append(action)
}
