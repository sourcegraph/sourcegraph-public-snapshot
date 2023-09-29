package codehost_scenario

import (
	"context"
	"fmt"

	"github.com/google/go-github/v53/github"
)

type Org struct {
	s    *GithubScenario
	name string
}

func (o *Org) Get(ctx context.Context) (*github.Organization, error) {
	if o.s.IsApplied() {
		return o.s.client.GetOrg(ctx, o.name)
	}
	panic("cannot retrieve org before scenario is applied")
}

func (o *Org) get(ctx context.Context) (*github.Organization, error) {
	return o.s.client.GetOrg(ctx, o.name)
}

func (o *Org) AllowPrivateForks() {
	updateOrgPermissions := &action{
		name: "org:permissions:update:" + o.name,
		apply: func(ctx context.Context) error {

			org, err := o.get(ctx)
			if err != nil {
				return err
			}
			org.MembersCanCreatePrivateRepos = boolp(true)
			org.MembersCanForkPrivateRepos = boolp(true)

			_, err = o.s.client.UpdateOrg(ctx, org)
			if err != nil {
				return err
			}
			return nil
		},
		teardown: nil,
	}
	o.s.append(updateOrgPermissions)
}

func (o *Org) CreateTeam(name string) any {
	createTeam := &action{
		name: "org:team:create:" + name,
		apply: func(ctx context.Context) error {
			return nil
		},
		teardown: func(ctx context.Context) error {
			return nil
		},
	}

	o.s.append(createTeam)

	return nil
}

func (o *Org) CreateRepo(name string, public bool) any {
	action := &action{
		name: fmt.Sprintf("repo:create:%s", name),
		apply: func(ctx context.Context) error {
			return nil
		},
		teardown: func(ctx context.Context) error {
			return nil
		},
	}
	o.s.append(action)

	return nil
}

func (o *Org) CreateRepoFork(target string) any {
	action := &action{
		name: fmt.Sprintf("repo:fork:%s", target),
		apply: func(ctx context.Context) error {
			return nil
		},
		teardown: func(ctx context.Context) error {
			return nil
		},
	}
	o.s.append(action)

	return nil
}
