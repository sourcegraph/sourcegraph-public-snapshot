package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/sourcegraph/log"
	"sigs.k8s.io/yaml"

	"github.com/sourcegraph/sourcegraph/dev/linearhooks/internal/lineargql"
	"github.com/sourcegraph/sourcegraph/dev/linearhooks/internal/linearschema"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	WildcardTeamID = "Any Team"
)

type Handler struct {
	logger log.Logger

	c             Config
	signingSecret string

	gqlClient graphql.Client
}

func New(ctx context.Context, logger log.Logger, b []byte, apiKey, signingSecret string) (*Handler, error) {
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}

	gqlClient := lineargql.NewGraphQLClient("https://api.linear.app/graphql", apiKey)

	return &Handler{
		logger:        logger,
		c:             c,
		signingSecret: signingSecret,
		gqlClient:     gqlClient,
	}, nil
}

func (h *Handler) HandleIssueMover(w http.ResponseWriter, r *http.Request) {
	body, err := h.validatePayloadSignature(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	eventType := r.Header.Get("Linear-Event")
	eventId := r.Header.Get("Linear-Delivery")
	logger := h.logger.With(log.String("event.type", eventType), log.String("event.id", eventId))
	if eventType != "Issue" && eventType != "Label" {
		logger.Warn("ignoring unexpected event type")
		return
	}

	var e linearschema.Event
	if err := json.Unmarshal(body, &e); err != nil {
		logger.Error("decode request body", log.Error(err))
		_, _ = w.Write([]byte("error decoding request body"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger = logger.With(log.String("event.action", string(e.Action)))

	if e.IssueData == nil {
		logger.Warn("missing issue data in event payload")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger = logger.With(log.String("issue.id", e.IssueData.Identifier))
	logger.Debug("processing issue data change event")

	switch e.Action {
	case linearschema.ActionTypeCreate, linearschema.ActionTypeUpdate:
	case linearschema.ActionTypeDelete:
		// no-op
		return
	default:
		logger.Warn("ignoring unexpected action type", log.String("action", string(e.Action)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := moveIssue(logger, h.gqlClient, h.c.Spec.Mover.Rules, *e.IssueData); err != nil {
		logger.Error("move issue", log.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte("OK"))
	w.WriteHeader(http.StatusOK)
}

func moveIssue(logger log.Logger, client graphql.Client, rules []RuleSpec, issue linearschema.IssueData) error {
	for _, r := range rules {
		if newTeamId := r.moveToNewTeam(issue); newTeamId != nil {
			logger.Info("moving issue",
				log.String("from.id", issue.Team.Key),
				log.String("to", r.Dst.TeamID),
			)
			teamUUID, err := getTeamUUID(context.Background(), client, r.Dst.TeamID)
			if err != nil {
				return errors.Wrapf(err, "get dst team %q UUID", r.Dst.TeamID)
			}
			if err := moveIssueToTeam(context.Background(), client, issue.Identifier, teamUUID); err != nil {
				return errors.Wrapf(err, "move issue %q to team %q", issue.Identifier, teamUUID)
			}
			return nil
		}
	}
	return nil
}

var (
	// teamKeyUUIDCache is a cache of team key to team UUID mappings with a 5 minute TTL.
	teamKeyUUIDCache = expirable.NewLRU[string, string](0, nil, time.Minute*5)
)

func getTeamUUID(ctx context.Context, c graphql.Client, idOrKey string) (string, error) {
	if v, ok := teamKeyUUIDCache.Get(idOrKey); ok {
		return v, nil
	}
	resp, err := lineargql.GetTeamById(ctx, c, idOrKey)
	if err != nil {
		return "", err
	}
	teamKeyUUIDCache.Add(idOrKey, resp.Team.Id)
	return resp.Team.Id, nil
}

func moveIssueToTeam(ctx context.Context, c graphql.Client, issueId, teamId string) error {
	_, err := lineargql.MoveIssueToTeam(ctx, c, issueId, teamId)
	return err
}

func (r *RuleSpec) moveToNewTeam(issueData linearschema.IssueData) (newTeamId *string) {
	if len(r.Src.Labels) == 0 {
		// This should never happen as we perform validation on startup, but if it does, we should panic.
		panic("Expected non-empty label set for RuleSet")
	}
	teamMatch := r.Src.TeamID == issueData.Team.ID || r.Src.TeamID == issueData.Team.Key || r.Src.TeamID == WildcardTeamID
	currentLabels := collections.NewSet(issueData.LabelNames()...)
	if teamMatch && collections.NewSet(r.Src.Labels...).Difference(currentLabels).IsEmpty() {
		return &r.Dst.TeamID
	}
	return nil
}
