package apitest

import (
	"fmt"
	"strings"
)

var Mutations = struct {
	DeleteCampaign          func(campaignID string) string
	AddChangesetsToCampaign func(countsFrom, countsTo, campaignID, changesetIDs string) string
}{
	DeleteCampaign: func(campaignID string) string {
		return fmt.Sprintf(`mutation { deleteCampaign(campaign: %q) { alwaysNil } }`, campaignID)
	},
	AddChangesetsToCampaign: func(countsFrom, countsTo, campaignID, changesetIDs string) string {
		fragments := []string{
			Fragments.User,
			Fragments.Org,
			Fragments.GitRef,
			Fragments.ExternalChangeset,
		}

		return fmt.Sprintf(strings.Join(fragments, "\n")+`
		fragment c on Campaign {
			id, name, description, createdAt, updatedAt
			author    { ...user }
			namespace {
				... on User { ...user }
				... on Org  { ...org }
			}
			changesets {
				nodes {
				  ... on ExternalChangeset {
				    ...externalChangeset
				  }
				}
				totalCount
				pageInfo { hasNextPage }
			}
			changesetCountsOverTime(from: %s, to: %s) {
			    date
				total
				merged
				closed
				open
				openApproved
				openChangesRequested
				openPending
			}
			diffStat {
				added
				changed
				deleted
			}
		}
		mutation() {
			campaign: addChangesetsToCampaign(campaign: %q, changesets: %s) {
				...c
			}
		}
		`, countsFrom, countsTo, campaignID, changesetIDs)
	},
}

var Fragments = struct {
	User               string
	Org                string
	SimpleCampaign     string
	SimpleCampaignConn string
	GitRef             string
	ExternalChangeset  string
	Patch              string
}{
	User: `fragment user on User { id, databaseID, siteAdmin }`,
	Org:  `fragment org on Org { id, name }`,
	SimpleCampaign: `
		fragment simpleCampaign on Campaign {
			id, name, description, createdAt, updatedAt
			author    { ...user }
			namespace {
				... on User { ...user }
				... on Org  { ...org }
			}
		}
	`,
	SimpleCampaignConn: `
		fragment simpleCampaignConn on CampaignConnection {
			nodes { ...simpleCampaign }
			totalCount
			pageInfo { hasNextPage }
		}
	`,
	GitRef: `
		fragment gitRef on GitRef {
			name
			abbrevName
			displayName
			prefix
			type
			repository { id }
			url
			target {
				oid
				abbreviatedOID
				type
			}
		}
		`,
	ExternalChangeset: `
		fragment externalChangeset on ExternalChangeset {
			id
			repository { id }
			createdAt
			updatedAt
			title
			body
			state
			externalURL {
				url
				serviceType
			}
			reviewState
			checkState
			events(first: 100) {
				totalCount
			}
			campaigns { nodes { id } }
			head { ...gitRef }
			base { ...gitRef }
		}
	`,
	Patch: `
		fragment patch on Patch {
            id
            publicationEnqueued
			repository {
				name
			}
			diff {
				fileDiffs {
					rawDiff
					diffStat {
						added
						deleted
						changed
					}
					nodes {
						oldPath
						newPath
						hunks {
							body
							section
							newRange { startLine, lines }
							oldRange { startLine, lines }
							oldNoNewlineAt
						}
						stat {
							added
							deleted
							changed
						}
						oldFile {
							name
							externalURLs {
							serviceType
							url
							}
						}
					}
				}
			}
		}
	`,
}
