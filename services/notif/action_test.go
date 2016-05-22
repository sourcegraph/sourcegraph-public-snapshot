package notif

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func TestGenerateMessage(t *testing.T) {
	cases := []struct {
		ActionContext ActionContext
		SlackMessage  string
		HTMLFragment  string
		EmailSubject  string
	}{
		{
			ActionContext{
				Person: &sourcegraph.Person{
					PersonSpec: sourcegraph.PersonSpec{
						Login: "keegancsmith",
					},
				},
				ActionType:  "created",
				ObjectRepo:  "sourcegraph",
				ObjectType:  "discussion",
				ObjectID:    6,
				ObjectTitle: "No rename in VFS",
				ObjectURL:   "https://sourcegraph.com/sourcegraph/sourcegraph/.discussion/6",
			},
			"*keegancsmith* created <https://sourcegraph.com/sourcegraph/sourcegraph/.discussion/6|sourcegraph discussion #6>: No rename in VFS",
			`<b>keegancsmith</b> created <a href="https://sourcegraph.com/sourcegraph/sourcegraph/.discussion/6">sourcegraph discussion #6</a>: No rename in VFS`,
			"[sourcegraph][Discussion #6] No rename in VFS",
		},
		{
			ActionContext{
				Person: &sourcegraph.Person{
					PersonSpec: sourcegraph.PersonSpec{
						Login: "sqs",
					},
				},
				ActionType:  "commented on",
				ObjectRepo:  "lib/annotate",
				ObjectType:  "discussion",
				ObjectID:    1,
				ObjectTitle: "What is the writeContent param for?",
				ObjectURL:   "https://sourcegraph.com/sourcegraph/lib/annotate/.discussion/1",
			},
			"*sqs* commented on <https://sourcegraph.com/sourcegraph/lib/annotate/.discussion/1|lib/annotate discussion #1>: What is the writeContent param for?",
			`<b>sqs</b> commented on <a href="https://sourcegraph.com/sourcegraph/lib/annotate/.discussion/1">lib/annotate discussion #1</a>: What is the writeContent param for?`,
			"[lib/annotate][Discussion #1] What is the writeContent param for?",
		},
		{
			ActionContext{
				Person: &sourcegraph.Person{
					PersonSpec: sourcegraph.PersonSpec{
						Login: "keegancsmith",
					},
				},
				Recipients: []*sourcegraph.Person{
					{
						PersonSpec: sourcegraph.PersonSpec{
							Login: "neelance",
						},
					},
				},
				ActionType:    "reviewed",
				ActionContent: "Ship it",
				ObjectRepo:    "sourcegraph",
				ObjectType:    "changeset",
				ObjectID:      71,
				ObjectTitle:   "Upgrade React to v0.14",
				ObjectURL:     "https://sourcegraph.com/sourcegraph/sourcegraph/.changesets/71",
			},
			"*keegancsmith* reviewed <https://sourcegraph.com/sourcegraph/sourcegraph/.changesets/71|sourcegraph changeset #71>: Upgrade React to v0.14 /cc @neelance\n\nShip it",
			`<b>keegancsmith</b> reviewed <a href="https://sourcegraph.com/sourcegraph/sourcegraph/.changesets/71">sourcegraph changeset #71</a>: Upgrade React to v0.14`,
			"[sourcegraph][Changeset #71] Upgrade React to v0.14",
		},
		{
			ActionContext{
				Person: &sourcegraph.Person{
					PersonSpec: sourcegraph.PersonSpec{
						Login: "renfredxh",
					},
				},
				ActionType:    "created",
				ActionContent: "Hi",
				ObjectRepo:    "lib/annotate",
				ObjectType:    "changeset",
				ObjectID:      2,
				ObjectTitle:   "Hello",
				ObjectURL:     "https://sourcegraph.com/sourcegraph/lib/annotate/.changesets/2",
			},
			"*renfredxh* created <https://sourcegraph.com/sourcegraph/lib/annotate/.changesets/2|lib/annotate changeset #2>: Hello\n\nHi",
			`<b>renfredxh</b> created <a href="https://sourcegraph.com/sourcegraph/lib/annotate/.changesets/2">lib/annotate changeset #2</a>: Hello`,
			"[lib/annotate][Changeset #2] Hello",
		},
		{
			// Test custom Slack message and email body.
			ActionContext{
				Person: &sourcegraph.Person{
					PersonSpec: sourcegraph.PersonSpec{
						Login: "pararthshah",
					},
				},
				ActionType:    "updated",
				ActionContent: "Hi",
				ObjectRepo:    "metrics",
				ObjectType:    "dashboard",
				ObjectID:      2,
				ObjectTitle:   "Hello",
				ObjectURL:     "https://sourcegraph.com/sourcegraph/metrics/.dashboard/2",
				SlackMsg:      "*pararthshah* updated metrics with a new <https://sourcegraph.com/sourcegraph/metrics/.dashboard/2|dashboard>",
				EmailHTML:     `<b>pararthshah</b> updated metrics with a new <a href="https://sourcegraph.com/sourcegraph/metrics/.dashboard/2">dashboard</a>`,
			},
			"*pararthshah* updated metrics with a new <https://sourcegraph.com/sourcegraph/metrics/.dashboard/2|dashboard>",
			`<b>pararthshah</b> updated metrics with a new <a href="https://sourcegraph.com/sourcegraph/metrics/.dashboard/2">dashboard</a>`,
			"[metrics][Dashboard #2] Hello",
		},
	}
	for _, c := range cases {
		msg, err := generateSlackMessage(c.ActionContext)
		if err != nil {
			t.Errorf("generateSlackMessage(%#v): %s", c.ActionContext, err)
		} else if msg != c.SlackMessage {
			t.Errorf("generateSlackMessage(%#v):\n%#v !=\n%#v", c.ActionContext, msg, c.SlackMessage)
		}

		msg, err = generateHTMLFragment(c.ActionContext)
		if err != nil {
			t.Errorf("generateHTMLFragment(%#v): %s", c.ActionContext, err)
		} else if msg != c.HTMLFragment {
			t.Errorf("generateHTMLFragment(%#v):\n%#v !=\n%#v", c.ActionContext, msg, c.HTMLFragment)
		}

		msg, err = generateEmailSubject(c.ActionContext)
		if err != nil {
			t.Errorf("generateEmailSubject(%#v): %s", c.ActionContext, err)
		} else if msg != c.EmailSubject {
			t.Errorf("generateEmailSubject(%#v):\n%#v !=\n%#v", c.ActionContext, msg, c.EmailSubject)
		}
	}
}
