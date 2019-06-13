package rules_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/rules"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDB_Rules(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	rules.ResetMocks()
	ctx := dbtesting.TestContext(t)

	// for testing equality of all other fields
	norm := func(vs ...*rules.DBRule) {
		for _, v := range vs {
			v.ID = 0
			v.CreatedAt = time.Time{}
			v.UpdatedAt = time.Time{}
		}
	}

	org1, err := db.Orgs.Create(ctx, "org1", nil)
	if err != nil {
		t.Fatal(err)
	}
	campaign1, err := campaigns.TestCreateCampaign(ctx, "p1", 0, org1.ID)
	if err != nil {
		t.Fatal(err)
	}
	campaign2, err := campaigns.TestCreateCampaign(ctx, "p2", 0, org1.ID)
	if err != nil {
		t.Fatal(err)
	}

	wantRule0 := &rules.DBRule{
		Container:   rules.RuleContainer{Campaign: campaign1},
		Name:        "n0",
		Description: strptr("d0"),
		Definition:  "h0",
	}
	rule0, err := rules.DBRules{}.Create(ctx, wantRule0)
	if err != nil {
		t.Fatal(err)
	}
	rule0ID := rule0.ID
	rule1, err := rules.DBRules{}.Create(ctx, &rules.DBRule{
		Container:   rules.RuleContainer{Campaign: campaign1},
		Name:        "n1",
		Description: strptr("d1"),
		Definition:  "h1",
	})
	if err != nil {
		t.Fatal(err)
	}
	{
		// Check Create result.
		if rule0.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		norm(rule0)
		if !reflect.DeepEqual(rule0, wantRule0) {
			t.Errorf("got %+v, want %+v", rule0, wantRule0)
		}
	}

	{
		// Get a rule.
		rule, err := rules.DBRules{}.GetByID(ctx, rule0ID)
		if err != nil {
			t.Fatal(err)
		}
		if rule.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		norm(rule)
		if !reflect.DeepEqual(rule, wantRule0) {
			t.Errorf("got %+v, want %+v", rule, wantRule0)
		}
	}

	{
		// List all rules.
		ts, err := rules.DBRules{}.List(ctx, rules.DBRulesListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		const want = 2
		if len(ts) != want {
			t.Errorf("got %d rules, want %d", len(ts), want)
		}
		count, err := rules.DBRules{}.Count(ctx, rules.DBRulesListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	}

	{
		// List campaign1's rules.
		ts, err := rules.DBRules{}.List(ctx, rules.DBRulesListOptions{Container: rules.RuleContainer{Campaign: campaign1}})
		if err != nil {
			t.Fatal(err)
		}
		const want = 2
		if len(ts) != want {
			t.Errorf("got %d rules, want %d", len(ts), want)
		}
	}

	{
		// List campaign2's rules.
		ts, err := rules.DBRules{}.List(ctx, rules.DBRulesListOptions{Container: rules.RuleContainer{Campaign: campaign2}})
		if err != nil {
			t.Fatal(err)
		}
		const want = 0
		if len(ts) != want {
			t.Errorf("got %d rules, want %d", len(ts), want)
		}
	}

	{
		// Query rules.
		ts, err := rules.DBRules{}.List(ctx, rules.DBRulesListOptions{Query: "n1"})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*rules.DBRule{rule1}; !reflect.DeepEqual(ts, want) {
			t.Errorf("got %+v, want %+v", ts, want)
		}
	}

	{
		// Delete a rule.
		if err := (rules.DBRules{}).DeleteByID(ctx, rule0ID); err != nil {
			t.Fatal(err)
		}
		ts, err := rules.DBRules{}.List(ctx, rules.DBRulesListOptions{Container: rules.RuleContainer{Campaign: campaign1}})
		if err != nil {
			t.Fatal(err)
		}
		const want = 1
		if len(ts) != want {
			t.Errorf("got %d rules, want %d", len(ts), want)
		}
	}
}

func strptr(s string) *string { return &s }
