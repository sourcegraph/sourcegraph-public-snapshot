package notif

import (
	"testing"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestGenerateMessage(t *testing.T) {
	cases := []struct {
		ActionContext ActionContext
		SlackMessage  string
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
				ObjectURL:   "https://src.sourcegraph.com/sourcegraph/.discussion/6",
			},
			"*keegancsmith* created <https://src.sourcegraph.com/sourcegraph/.discussion/6|sourcegraph discussion #6>: No rename in VFS",
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
				ObjectURL:   "https://src.sourcegraph.com/lib/annotate/.discussion/1",
			},
			"*sqs* commented on <https://src.sourcegraph.com/lib/annotate/.discussion/1|lib/annotate discussion #1>: What is the writeContent param for?",
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
				ObjectURL:     "https://src.sourcegraph.com/sourcegraph/.changesets/71",
			},
			"*keegancsmith* reviewed <https://src.sourcegraph.com/sourcegraph/.changesets/71|sourcegraph changeset #71>: Upgrade React to v0.14 /cc @neelance\n\nShip it",
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
				ObjectURL:     "https://src.sourcegraph.com/lib/annotate/.changesets/2",
			},
			"*renfredxh* created <https://src.sourcegraph.com/lib/annotate/.changesets/2|lib/annotate changeset #2>: Hello\n\nHi",
		},
	}
	for _, c := range cases {
		msg, err := generateSlackMessage(c.ActionContext)
		if err != nil {
			t.Errorf("generateSlackMessage(%#v): %s", c.ActionContext, err)
		} else if msg != c.SlackMessage {
			t.Errorf("generateSlackMessage(%#v):\n%#v !=\n%#v", c.ActionContext, msg, c.SlackMessage)
		}
	}
}
