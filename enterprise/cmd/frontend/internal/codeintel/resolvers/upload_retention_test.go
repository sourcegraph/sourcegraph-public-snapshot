package resolvers

import (
	"context"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestRetentionPolicyOverview(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()

	resolver := NewResolver(mockDBStore, mockLSIFStore, mockGitserverClient, nil, nil, nil, nil, 50, &observation.TestContext, nil)

	mockClock := glock.NewMockClock()

	cases := []struct {
		name            string
		expectedMatches int
		upload          dbstore.Upload
		mockPolicies    []RetentionPolicyMatchCandidate
		refDescriptions map[string][]gitdomain.RefDescription
	}{
		{
			name:            "basic single upload match",
			expectedMatches: 1,
			upload: dbstore.Upload{
				Commit:     "deadbeef0",
				UploadedAt: mockClock.Now().Add(-time.Hour * 23),
			},
			mockPolicies: []RetentionPolicyMatchCandidate{
				{
					ConfigurationPolicy: &dbstore.ConfigurationPolicy{
						RetentionDuration:         timePtr(time.Hour * 24),
						RetainIntermediateCommits: false,
						Type:                      dbstore.GitObjectTypeTag,
						Pattern:                   "*",
					},
					Matched: true,
				},
			},
			refDescriptions: map[string][]gitdomain.RefDescription{
				"deadbeef0": {
					{
						Name:            "v4.2.0",
						Type:            gitdomain.RefTypeTag,
						IsDefaultBranch: false,
					},
				},
			},
		},
		{
			name:            "matching but expired",
			expectedMatches: 0,
			upload: dbstore.Upload{
				Commit:     "deadbeef0",
				UploadedAt: mockClock.Now().Add(-time.Hour * 25),
			},
			mockPolicies: []RetentionPolicyMatchCandidate{
				{
					ConfigurationPolicy: &dbstore.ConfigurationPolicy{
						RetentionDuration:         timePtr(time.Hour * 24),
						RetainIntermediateCommits: false,
						Type:                      dbstore.GitObjectTypeTag,
						Pattern:                   "*",
					},
					Matched: false,
				},
			},
			refDescriptions: map[string][]gitdomain.RefDescription{
				"deadbeef0": {
					{
						Name:            "v4.2.0",
						Type:            gitdomain.RefTypeTag,
						IsDefaultBranch: false,
					},
				},
			},
		},
		{
			name:            "tip of default branch match",
			expectedMatches: 1,
			upload: dbstore.Upload{
				Commit:     "deadbeef0",
				UploadedAt: mockClock.Now().Add(-time.Hour * 25),
			},
			mockPolicies: []RetentionPolicyMatchCandidate{
				{
					ConfigurationPolicy: nil,
					Matched:             true,
				},
			},
			refDescriptions: map[string][]gitdomain.RefDescription{
				"deadbeef0": {
					{
						Name:            "main",
						Type:            gitdomain.RefTypeBranch,
						IsDefaultBranch: true,
					},
				},
			},
		},
		{
			name:            "direct match (1 of 2 policies)",
			expectedMatches: 1,
			upload: dbstore.Upload{
				Commit:     "deadbeef0",
				UploadedAt: mockClock.Now().Add(-time.Minute),
			},
			mockPolicies: []RetentionPolicyMatchCandidate{
				{
					ConfigurationPolicy: &dbstore.ConfigurationPolicy{
						RetentionDuration:         timePtr(time.Hour * 24),
						RetainIntermediateCommits: false,
						Type:                      dbstore.GitObjectTypeTag,
						Pattern:                   "*",
					},
					Matched: true,
				},
				{
					ConfigurationPolicy: &dbstore.ConfigurationPolicy{
						RetentionDuration:         timePtr(time.Hour * 24),
						RetainIntermediateCommits: false,
						Type:                      dbstore.GitObjectTypeTree,
						Pattern:                   "*",
					},
					Matched: false,
				},
			},
			refDescriptions: map[string][]gitdomain.RefDescription{
				"deadbeef0": {
					{
						Name:            "v4.2.0",
						Type:            gitdomain.RefTypeTag,
						IsDefaultBranch: false,
					},
				},
			},
		},
		{
			name:            "direct match (ignore visible)",
			expectedMatches: 1,
			upload: dbstore.Upload{
				Commit:     "deadbeef1",
				UploadedAt: mockClock.Now().Add(-time.Minute),
			},
			mockPolicies: []RetentionPolicyMatchCandidate{
				{
					ConfigurationPolicy: &dbstore.ConfigurationPolicy{
						RetentionDuration:         timePtr(time.Hour * 24),
						RetainIntermediateCommits: false,
						Type:                      dbstore.GitObjectTypeTag,
						Pattern:                   "*",
					},
					Matched: true,
				},
			},
			refDescriptions: map[string][]gitdomain.RefDescription{
				"deadbeef1": {
					{
						Name:            "v4.2.0",
						Type:            gitdomain.RefTypeTag,
						IsDefaultBranch: false,
					},
				},
				"deadbeef0": {
					{
						Name:            "v4.1.9",
						Type:            gitdomain.RefTypeTag,
						IsDefaultBranch: false,
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run("PolicyOverview "+c.name, func(t *testing.T) {
			expectedPolicyCandidates, mockedStorePolicies := mockConfigurationPolicies(c.mockPolicies)
			mockDBStore.GetConfigurationPoliciesFunc.PushReturn(mockedStorePolicies, len(mockedStorePolicies), nil)

			mockGitserverClient.RefDescriptionsFunc.PushReturn(c.refDescriptions, nil)

			matches, _, err := resolver.RetentionPolicyOverview(context.Background(), c.upload, false, 10, 0, "", mockClock.Now())
			if err != nil {
				t.Fatalf("unexpected error resolving retention policy overview: %v", err)
			}

			var matchCount int
			for _, match := range matches {
				if match.Matched {
					matchCount++
				}
			}

			if matchCount != c.expectedMatches {
				t.Errorf("unexpected number of matched policies: want=%d have=%d", c.expectedMatches, matchCount)
			}

			if diff := cmp.Diff(expectedPolicyCandidates, matches); diff != "" {
				t.Errorf("unexpected retention policy matches (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRetentionPolicyOverview_ByVisibility(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()

	resolver := NewResolver(mockDBStore, mockLSIFStore, mockGitserverClient, nil, nil, nil, nil, 50, &observation.TestContext, nil)

	mockClock := glock.NewMockClock()

	// deadbeef2 ----\
	// deadbeef0 ---- deadbeef1
	// T0------------------->T1

	cases := []struct {
		name            string
		upload          dbstore.Upload
		mockPolicies    []RetentionPolicyMatchCandidate
		visibleCommits  []string
		refDescriptions map[string][]gitdomain.RefDescription
		expectedMatches int
	}{
		{
			name:            "basic single visibility",
			expectedMatches: 1,
			upload: dbstore.Upload{
				Commit:     "deadbeef0",
				UploadedAt: mockClock.Now().Add(-time.Minute * 24),
			},
			visibleCommits: []string{"deadbeef1"},
			mockPolicies: []RetentionPolicyMatchCandidate{
				{
					ConfigurationPolicy: &dbstore.ConfigurationPolicy{
						RetentionDuration:         timePtr(time.Hour * 24),
						RetainIntermediateCommits: false,
						Type:                      dbstore.GitObjectTypeTag,
						Pattern:                   "*",
					},
					ProtectingCommits: []string{"deadbeef1"},
					Matched:           true,
				},
			},
			refDescriptions: map[string][]gitdomain.RefDescription{
				"deadbeef1": {
					{
						Name:            "v4.2.0",
						Type:            gitdomain.RefTypeTag,
						IsDefaultBranch: false,
					},
				},
			},
		},
		{
			name:            "visibile to tip of default branch",
			expectedMatches: 1,
			visibleCommits:  []string{"deadbeef0", "deadbeef1"},
			upload: dbstore.Upload{
				Commit:     "deadbeef0",
				UploadedAt: mockClock.Now().Add(-time.Hour * 24),
			},
			mockPolicies: []RetentionPolicyMatchCandidate{
				{
					ConfigurationPolicy: nil,
					ProtectingCommits:   []string{"deadbeef1"},
					Matched:             true,
				},
			},
			refDescriptions: map[string][]gitdomain.RefDescription{
				"deadbeef1": {
					{
						Name:            "main",
						Type:            gitdomain.RefTypeBranch,
						IsDefaultBranch: true,
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run("ByVisibility "+c.name, func(t *testing.T) {
			expectedPolicyCandidates, mockedStorePolicies := mockConfigurationPolicies(c.mockPolicies)
			mockDBStore.GetConfigurationPoliciesFunc.PushReturn(mockedStorePolicies, len(mockedStorePolicies), nil)
			mockDBStore.CommitsVisibleToUploadFunc.PushReturn(c.visibleCommits, nil, nil)

			mockGitserverClient.RefDescriptionsFunc.PushReturn(c.refDescriptions, nil)

			matches, _, err := resolver.RetentionPolicyOverview(context.Background(), c.upload, false, 10, 0, "", mockClock.Now())
			if err != nil {
				t.Fatalf("unexpected error resolving retention policy overview: %v", err)
			}

			var matchCount int
			for _, match := range matches {
				if match.Matched {
					matchCount++
				}
			}

			if matchCount != c.expectedMatches {
				t.Errorf("unexpected number of matched policies: want=%d have=%d", c.expectedMatches, matchCount)
			}

			if diff := cmp.Diff(expectedPolicyCandidates, matches); diff != "" {
				t.Errorf("unexpected retention policy matches (-want +got):\n%s", diff)
			}
		})
	}
}

func timePtr(t time.Duration) *time.Duration {
	return &t
}

func mockConfigurationPolicies(policies []RetentionPolicyMatchCandidate) (mockedCandidates []RetentionPolicyMatchCandidate, mockedPolicies []dbstore.ConfigurationPolicy) {
	for i, policy := range policies {
		if policy.ConfigurationPolicy != nil {
			policy.ID = i + 1
			mockedPolicies = append(mockedPolicies, *policy.ConfigurationPolicy)
		}
		policies[i] = policy
		mockedCandidates = append(mockedCandidates, policy)
	}
	return
}
