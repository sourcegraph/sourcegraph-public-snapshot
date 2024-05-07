package handlers

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/dev/linearhooks/internal/lineargql"
	"github.com/sourcegraph/sourcegraph/dev/linearhooks/internal/lineargql/gqltest"
	"github.com/sourcegraph/sourcegraph/dev/linearhooks/internal/linearschema"
)

func Test_moveIssue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		rs             []RuleSpec
		issue          linearschema.IssueData
		stub           gqltest.MakeRequestStub
		gqlCallCount   int
		gqlCallPayload autogold.Value
	}{
		{
			name: "no match",
			rs: []RuleSpec{
				{
					Src: SrcSpec{TeamID: "team-uuid", Labels: []string{"a"}},
					Dst: DstSpec{TeamID: "another-team-uuid"},
				},
			},
			issue: linearschema.IssueData{
				Team: linearschema.IssueTeamData{
					ID:  "other-team-uuid",
					Key: "other-team-key",
				},
			},
		},
		{
			name: "match",
			rs: []RuleSpec{
				{
					Src: SrcSpec{TeamID: "team-uuid", Labels: []string{"a"}},
					Dst: DstSpec{TeamID: "another-team-uuid"},
				},
			},
			issue: linearschema.IssueData{
				Identifier: "team-key-1",
				Labels:     []linearschema.IssueLabelData{{Name: "a"}},
				Team: linearschema.IssueTeamData{
					ID:  "team-uuid",
					Key: "team-key",
				},
			},
			stub: gqltest.MakeRequestStubInvocations(
				gqltest.MakeRequestResultStub(lineargql.GetTeamByIdResponse{Team: lineargql.GetTeamByIdTeam{Id: "another-team-uuid", Key: "another-team-key", Name: "another-team-name"}}),
				gqltest.MakeRequestResultStub(lineargql.MoveIssueToTeamResponse{}),
			),
			gqlCallCount: 2,
			gqlCallPayload: autogold.Expect([]map[string]interface{}{
				{"id": "another-team-uuid"},
				{
					"issueId": "team-key-1",
					"teamId":  "another-team-uuid",
				},
			}),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, r := range tc.rs {
				require.NoError(t, r.Validate())
			}

			faker := &gqltest.FakeClient{}
			if tc.stub != nil {
				faker.MakeRequestStub = tc.stub
			}
			err := moveIssue(logtest.Scoped(t), faker, tc.rs, tc.issue)
			require.NoError(t, err)
			assert.Equal(t, tc.gqlCallCount, faker.MakeRequestCallCount())

			if tc.gqlCallPayload != nil {
				var m []map[string]any
				for i := 0; i < tc.gqlCallCount; i++ {
					mm := gqltest.UnmarshalVariables(t, faker.MakeRequestArgsGraphqlRequestForCall(i))
					m = append(m, mm)
				}
				tc.gqlCallPayload.Equal(t, m)
			}
		})
	}
}

func Test_moveToNewTeam(t *testing.T) {
	t.Parallel()

	defaultDstSpec := DstSpec{TeamID: "dst-team-uuid"}

	tests := []struct {
		name         string
		r            RuleSpec
		issueTeamID  string
		issueTeamKey string
		issueLabels  []string
		want         bool
	}{
		{
			name:        "yes - full labels match with team id",
			r:           RuleSpec{Src: SrcSpec{TeamID: "team-id", Labels: []string{"a", "b"}}, Dst: defaultDstSpec},
			issueTeamID: "team-id",
			issueLabels: []string{"a", "b"},
			want:        true,
		},
		{
			name:         "yes - full labels match with team key",
			r:            RuleSpec{Src: SrcSpec{TeamID: "team-key", Labels: []string{"a", "b"}}, Dst: defaultDstSpec},
			issueTeamKey: "team-key",
			issueLabels:  []string{"a", "b"},
			want:         true,
		},
		{
			name:        "yes - desired labels are subset of the issue labels",
			r:           RuleSpec{Src: SrcSpec{TeamID: "team-id", Labels: []string{"a", "b"}}, Dst: defaultDstSpec},
			issueTeamID: "team-id",
			issueLabels: []string{"a", "b", "c"},
			want:        true,
		},
		{
			name:        "yes - wildcard team id",
			r:           RuleSpec{Src: SrcSpec{TeamID: WildcardTeamID, Labels: []string{"a", "b"}}, Dst: defaultDstSpec},
			issueTeamID: "random-team-id",
			issueLabels: []string{"a", "b"},
			want:        true,
		},
		{
			name:        "no - none of the labels match",
			r:           RuleSpec{Src: SrcSpec{TeamID: "team-id", Labels: []string{"c", "d"}}, Dst: defaultDstSpec},
			issueTeamID: "team-id",
			issueLabels: []string{"a", "b"},
		},
		{
			name:        "no - issue labels are subset of the desired labels",
			r:           RuleSpec{Src: SrcSpec{TeamID: "team-id", Labels: []string{"a", "b", "c"}}, Dst: defaultDstSpec},
			issueTeamID: "team-id",
			issueLabels: []string{"a", "b"},
		},
		{
			name:        "no - issue has no labels",
			r:           RuleSpec{Src: SrcSpec{TeamID: "team-id", Labels: []string{"a", "b", "c"}}, Dst: defaultDstSpec},
			issueTeamID: "team-id",
			issueLabels: []string{""},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.NoError(t, tc.r.Validate())
			got := tc.r.moveToNewTeam(
				linearschema.IssueData{
					Team:   linearschema.IssueTeamData{ID: tc.issueTeamID, Key: tc.issueTeamKey},
					Labels: labelsToLabelData(tc.issueLabels),
				},
			)
			if !tc.want {
				assert.Nil(t, got)
				return
			}
			assert.NotNil(t, got)
		})
	}
}

func labelsToLabelData(labels []string) []linearschema.IssueLabelData {
	var lds []linearschema.IssueLabelData
	for _, l := range labels {
		lds = append(lds, linearschema.IssueLabelData{Name: l})
	}
	return lds
}
