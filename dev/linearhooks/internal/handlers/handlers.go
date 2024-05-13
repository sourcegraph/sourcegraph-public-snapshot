package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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
		if newTeamId := r.identifyTeamToMoveTo(issue); newTeamId != nil {
			logger := logger.With(
				log.String("issue.id", issue.Identifier),
				log.String("from.key", issue.Team.Key),
				log.String("to.id", r.Dst.TeamID),
			)
			logger.Info("moving issue")

			teamUUID, err := getTeamUUID(context.Background(), client, r.Dst.TeamID)
			if err != nil {
				return errors.Wrapf(err, "get dst team %q UUID", r.Dst.TeamID)
			}

			input := moveIssueToTeamInput{
				issueId:  issue.Identifier,
				teamUUID: teamUUID,
			}
			if r.Dst.Modifier != nil && r.Dst.Modifier.ProjectName != "" {
				logger := logger.With(log.String("to.projectName", r.Dst.Modifier.ProjectName))
				projectUUID, err := getProjectUUID(context.Background(), client, teamUUID, r.Dst.Modifier.ProjectName)
				if err != nil {
					logger.Error("unable to resolve project UUID, issue is moved without setting the desired project. you should inspect the error and consider updating the rule.",
						log.Error(err),
						log.String("to.uuid", teamUUID),
					)
				} else {
					input.projectId = projectUUID
				}
			}
			if err := moveIssueToTeam(context.Background(), client, input); err != nil {
				return errors.Wrapf(err, "move issue %q to team %q", issue.Identifier, teamUUID)
			}
			return nil
		}
	}
	return nil
}

var (
	// teamKeyUUIDCache caches team identifier/key -> UUID mappings to avoid
	// re-requesting the same data for every label change.
	// The cache has a TTL for eventually correct behavior in case a team is
	// deleted and re-created.
	teamKeyUUIDCache = expirable.NewLRU[string, string](0, nil, time.Minute*5)

	// projectNameAndTeamIDToUUIDCache caches (team_uuid, project_name) -> UUID mappings to avoid
	// re-requesting the same data for every label change. It's quite expensive to list all projects.
	// The cache has a TTL for eventually correct behavior in case a project is
	// deleted and re-created.
	projectNameAndTeamIDToUUIDCache = expirable.NewLRU[string, string](0, nil, time.Minute*5)
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

func getProjectUUID(ctx context.Context, c graphql.Client, teamUUID, projectName string) (string, error) {
	compositeKey := fmt.Sprintf("%s:%s", teamUUID, projectName)
	if v, ok := projectNameAndTeamIDToUUIDCache.Get(compositeKey); ok {
		return v, nil
	}
	projects, err := lineargql.GetProjectsByTeamId(ctx, c, teamUUID, projectName)
	if err != nil {
		return "", errors.Wrapf(err, "list projects by team ID %s", teamUUID)
	}
	for _, project := range projects.Team.Projects.Nodes {
		if project.Name == projectName {
			projectNameAndTeamIDToUUIDCache.Add(compositeKey, project.Id)
			return project.Id, nil
		}
	}
	return "", errors.Newf("project %s not found", projectName)
}

type moveIssueToTeamInput struct {
	issueId   string
	teamUUID  string
	projectId string
}

func moveIssueToTeam(ctx context.Context, c graphql.Client, input moveIssueToTeamInput) error {
	_, err := lineargql.MoveIssueToTeam(ctx, c, input.issueId, input.teamUUID, input.projectId)
	return err
}

func (r *RuleSpec) identifyTeamToMoveTo(issueData linearschema.IssueData) (newTeamId *string) {
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
