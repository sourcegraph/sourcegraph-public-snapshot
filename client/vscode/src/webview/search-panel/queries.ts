import { gql } from '@sourcegraph/shared/src/graphql/graphql'

export const searchQuery = gql`
    query Search($query: String!, $patternType: SearchPatternType) {
        search(version: V2, query: $query, patternType: $patternType) {
            results {
                matchCount
                repositoriesCount
                elapsedMilliseconds
                limitHit
                results {
                    __typename
                    ... on FileMatch {
                        ...FileMatchFields
                    }
                    ... on CommitSearchResult {
                        ...CommitSearchResultFields
                    }
                    ... on Repository {
                        description
                        ...RepositoryFields
                    }
                }
                ...SearchResultsAlertFields
                ...DynamicFiltersFields
            }
        }
    }
    fragment FileMatchFields on FileMatch {
        repository {
            ...RepositoryFields
        }
        file {
            name
            path
            url
            content
            commit {
                oid
            }
        }
        lineMatches {
            preview
            lineNumber
            offsetAndLengths
        }
        symbols {
            url
            name
            containerName
            kind
        }
    }
    fragment CommitSearchResultFields on CommitSearchResult {
        label {
            text
        }
        detail {
            text
        }
        commit {
            url
            message
            author {
                person {
                    name
                }
            }
            repository {
                ...RepositoryFields
            }
        }
        matches {
            url
            body {
                text
            }
            highlights {
                character
                line
                length
            }
        }
    }
    fragment RepositoryFields on Repository {
        name
        url
        stars
    }
    fragment SearchResultsAlertFields on SearchResults {
        alert {
            title
            description
            proposedQueries {
                description
                query
            }
        }
    }
    fragment DynamicFiltersFields on SearchResults {
        dynamicFilters {
            value
            label
            count
            limitHit
            kind
        }
    }
`

export const fileNamesQuery = gql`
    query FileNames($repository: String!, $revision: String!) {
        repository(name: $repository) {
            commit(rev: $revision) {
                fileNames
            }
        }
    }
`

const savedSearchFragment = gql`
    fragment SavedSearchFields on SavedSearch {
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
`

export const createSavedSearchQuery = gql`
    mutation CreateSavedSearch(
        $description: String!
        $query: String!
        $notifyOwner: Boolean!
        $notifySlack: Boolean!
        $userID: ID
        $orgID: ID
    ) {
        createSavedSearch(
            description: $description
            query: $query
            notifyOwner: $notifyOwner
            notifySlack: $notifySlack
            userID: $userID
            orgID: $orgID
        ) {
            ...SavedSearchFields
        }
    }
    ${savedSearchFragment}
`
export const treeEntriesQuery = gql`
    query TreeEntries($repoName: String!, $revision: String!, $commitID: String!, $filePath: String!, $first: Int) {
        repository(name: $repoName) {
            commit(rev: $commitID, inputRevspec: $revision) {
                tree(path: $filePath) {
                    ...TreeFields
                }
            }
        }
    }
    fragment TreeFields on GitTree {
        isRoot
        url
        entries(first: $first, recursiveSingleChild: true) {
            ...TreeEntryFields
        }
    }
    fragment TreeEntryFields on TreeEntry {
        name
        path
        isDirectory
        url
        submodule {
            url
            commit
        }
        isSingleChild
    }
`

export const eventsQuery = gql`
    query EventLogsData($userId: ID!, $first: Int, $eventName: String!) {
        node(id: $userId) {
            ... on User {
                __typename
                eventLogs(first: $first, eventName: $eventName) {
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
        }
    }
`
export const currentAuthStateQuery = gql`
    query CurrentAuthState {
        currentUser {
            __typename
            id
            databaseID
            username
            avatarURL
            email
            displayName
            siteAdmin
            tags
            url
            settingsURL
            organizations {
                nodes {
                    id
                    name
                    displayName
                    url
                    settingsURL
                }
            }
            session {
                canSignOut
            }
            viewerCanAdminister
            tags
        }
    }
`
