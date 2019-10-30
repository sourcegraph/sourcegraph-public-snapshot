package run

import (
	"errors"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
)

var combyServiceURL string

func init() {
	combyServiceURL = env.Get("COMBY_URL", "http://replacer:3185", "replacer server URL")
}

type combyArgs struct {
	ScopeQuery      string `json:"scopeQuery"`
	MatchTemplate   string `json:"matchTemplate"`
	RewriteTemplate string `json:"rewriteTemplate"`
}

type comby struct {
	plan *a8n.CampaignPlan

	args combyArgs
}

func (c *comby) validateArgs() error {
	if err := jsonc.Unmarshal(c.plan.Arguments, &c.args); err != nil {
		return err
	}

	if c.args.ScopeQuery == "" {
		return errors.New("missing argument in specification: scopeQuery")
	}

	if c.args.MatchTemplate == "" {
		return errors.New("missing argument in specification: matchTemplate")
	}

	if c.args.RewriteTemplate == "" {
		return errors.New("missing argument in specification: rewriteTemplate")
	}

	return nil
}

func (c *comby) searchQuery() string { return c.args.ScopeQuery }
func (c *comby) runJob(j *a8n.CampaignJob) {
	// TODO(a8n): Do real work.
	j.Diff = bogusDiff
	j.Error = ""
	j.FinishedAt = time.Now()
}

const bogusDiff = `diff --git a/README.md b/README.md
index 323fae0..34a3ec2 100644
--- a/README.md
+++ b/README.md
@@ -1 +1 @@
-foobar
+barfoo
`
