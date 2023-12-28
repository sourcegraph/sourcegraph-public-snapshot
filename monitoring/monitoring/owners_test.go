package monitoring

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/hexops/autogold/v2"
	opsgenie "github.com/opsgenie/opsgenie-go-sdk-v2/client"
	opsgenieteam "github.com/opsgenie/opsgenie-go-sdk-v2/team"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var onlineCheck = flag.Bool("online", false, "Run online checks")

type opsgenieResponderConfig struct {
	Type string `json:"type"` // "team"
	Name string `json:"name"` // owner.opsgenieTeam
}

type opsgenieNotifierConfig struct {
	Type       string                    `json:"type"` // "opgsenie"
	Responders []opsgenieResponderConfig `json:"responders"`
}

type notifierConfig struct {
	Level    string                 `json:"level"` // "critical"
	Notifier opsgenieNotifierConfig `json:"notifier"`
	Owners   []string               `json:"owners"` // owner.opsgenieTeam
}

// TestOwnersOpsgenieTeam checks Opsgenie team details of each owner.
func TestOwnersOpsgenieTeam(t *testing.T) {
	if !*onlineCheck {
		t.Skip("MONITORING_OWNERS_ONLINE_CHECK not set to true, skipping online checks")
	}

	opsgenieKey := os.Getenv("OPSGENIE_API_KEY")
	if opsgenieKey == "" {
		t.Fatal("OPSGENIE_API_KEY not set, skipping test")
	}

	client, err := opsgenieteam.NewClient(&opsgenie.Config{
		ApiKey: opsgenieKey,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// As part of this test, we also build notifier config of all valid owners
	// so that they can be included in Sourcegraph.com's 'observability.alerts'
	// configuration. Configuration with invalid targets means that alerts might
	// end up not going to _any_ team, so we want to make sure to skip those
	// owners. If a team is kind enough to set up a real owner and opsgenie team
	// this helps make sure they are included as well.
	var observabilityAlertsConfig []notifierConfig

	var failed int
	for _, owner := range allKnownOwners {
		if t.Run(owner.teamName, func(t *testing.T) {
			team, err := client.Get(ctx, &opsgenieteam.GetTeamRequest{
				IdentifierType:  opsgenieteam.Name,
				IdentifierValue: owner.opsgenieTeam,
			})
			assert.NoError(t, err)
			if assert.NotNil(t, team) {
				t.Logf("TeamMeta: %+v", team.TeamMeta)
				t.Logf("Description: %q", team.Description)
				t.Logf("Members: %d", len(team.Members))
			}
		}) {
			observabilityAlertsConfig = append(observabilityAlertsConfig, notifierConfig{
				Level: "critical",
				Notifier: opsgenieNotifierConfig{
					Type: "opsgenie",
					Responders: []opsgenieResponderConfig{{
						Type: "team",
						Name: owner.opsgenieTeam,
					}},
				},
				Owners: []string{owner.opsgenieTeam},
			})
		} else {
			failed += 1
		}
	}

	var data bytes.Buffer
	enc := json.NewEncoder(&data)
	enc.SetIndent("    ", "  ")
	assert.NoError(t, enc.Encode(observabilityAlertsConfig))
	// The below can be copy-pasted into
	// https://sourcegraph.sourcegraph.com/search?q=context:global+repo:github.com/sourcegraph/deploy-sourcegraph-cloud+file:overlays/prod/frontend/files/site.json+%22observability.alerts%22:+%5B...%5D&patternType=structural&sm=1&groupBy=repo
	autogold.Expect(`[
      {
        "level": "critical",
        "notifier": {
          "type": "opsgenie",
          "responders": [
            {
              "type": "team",
              "name": "infra-support"
            }
          ]
        },
        "owners": [
          "infra-support"
        ]
      },
      {
        "level": "critical",
        "notifier": {
          "type": "opsgenie",
          "responders": [
            {
              "type": "team",
              "name": "code-insights"
            }
          ]
        },
        "owners": [
          "code-insights"
        ]
      },
      {
        "level": "critical",
        "notifier": {
          "type": "opsgenie",
          "responders": [
            {
              "type": "team",
              "name": "code-intel"
            }
          ]
        },
        "owners": [
          "code-intel"
        ]
      },
      {
        "level": "critical",
        "notifier": {
          "type": "opsgenie",
          "responders": [
            {
              "type": "team",
              "name": "source"
            }
          ]
        },
        "owners": [
          "source"
        ]
      }
    ]
`).Equal(t, data.String())

	if failed > 0 {
		t.Errorf("%d/%d ObservableOwners do not have valid Opsgenie teams",
			failed, len(allKnownOwners))
	}
}

// TestOwnersHandbookPages checks if the handbook page URLs of each owner is
// valid and exists.
func TestOwnersHandbookPages(t *testing.T) {
	if !*onlineCheck {
		t.Skip("MONITORING_OWNERS_ONLINE_CHECK not set to true, skipping online checks")
	}

	var failed int
	for _, owner := range allKnownOwners {
		if !t.Run(owner.teamName, func(t *testing.T) {
			page, err := url.Parse(owner.getHandbookPageURL())
			if !assert.NoError(t, err) {
				return
			}

			resp, err := http.DefaultClient.Do(&http.Request{
				Method: http.MethodGet,
				URL:    page,
			})
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}) {
			failed += 1
		}
	}

	if failed > 0 {
		t.Errorf("%d/%d ObservableOwners do not point to valid handbook pages",
			failed, len(allKnownOwners))
	}
}
