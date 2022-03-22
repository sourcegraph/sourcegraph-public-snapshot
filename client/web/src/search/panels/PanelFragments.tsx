import { gql } from '@sourcegraph/http-client'

// We define all fragments for the different homepage panels in this file to
// avoid creating circular imports in Jest, see:
//
//  - https://blackdeerdev.com/graphqlerror-syntax-error-unexpected-name-undefined/
//  - https://spectrum.chat/apollo/general/fragments-not-working-cross-files-in-mutation~c4e90568-f89a-458f-9810-0730846fc5f0

export const collaboratorsFragment = gql`
    fragment CollaboratorsFragment on User {
        collaborators: invitableCollaborators @include(if: $enableCollaborators) {
            name
            email
            displayName
            avatarURL
        }
    }
`

export const recentFilesFragment = gql`
    fragment RecentFilesFragment on User {
        recentFilesLogs: eventLogs(first: $firstRecentFiles, eventName: "ViewBlob") {
            nodes {
                argument
                timestamp
                url
            }
            pageInfo {
                hasNextPage
            }
            totalCount
        }
    }
`

export const recentSearchesPanelFragment = gql`
    fragment RecentSearchesPanelFragment on User {
        recentSearchesLogs: eventLogs(first: $firstRecentSearches, eventName: "SearchResultsQueried") {
            nodes {
                argument
                timestamp
                url
            }
            pageInfo {
                hasNextPage
            }
            totalCount
        }
    }
`

export const recentlySearchedRepositoriesFragment = gql`
    fragment RecentlySearchedRepositoriesFragment on User {
        recentlySearchedRepositoriesLogs: eventLogs(
            first: $firstRecentlySearchedRepositories
            eventName: "SearchResultsQueried"
        ) {
            nodes {
                argument
                timestamp
                url
            }
            pageInfo {
                hasNextPage
            }
            totalCount
        }
    }
`

export const savedSearchesPanelFragment = gql`
    fragment SavedSearchesPanelFragment on Query {
        savedSearches @include(if: $enableSavedSearches) {
            id
            description
            notify
            notifySlack
            query
            namespace {
                __typename
                id
                namespaceName
            }
            slackWebhookURL
        }
    }
`
