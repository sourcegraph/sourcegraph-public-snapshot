import { gql } from '@sourcegraph/http-client'

const savedSearchFragment = gql`
    fragment SavedSearchFields on SavedSearch {
        id
        description
        query
        draft
        owner {
            __typename
            id
            namespaceName
        }
        visibility
        createdAt
        createdBy {
            id
            username
            url
        }
        updatedAt
        updatedBy {
            id
            username
            url
        }
        url
        viewerCanAdminister
    }
`

export const savedSearchesQuery = gql`
    query SavedSearches(
        $query: String
        $owner: ID = null
        $viewerIsAffiliated: Boolean
        $includeDrafts: Boolean = true
        $first: Int
        $last: Int
        $after: String
        $before: String
        $orderBy: SavedSearchesOrderBy!
    ) {
        savedSearches(
            query: $query
            owner: $owner
            viewerIsAffiliated: $viewerIsAffiliated
            includeDrafts: $includeDrafts
            first: $first
            last: $last
            after: $after
            before: $before
            orderBy: $orderBy
        ) {
            nodes {
                ...SavedSearchFields
            }
            totalCount
            pageInfo {
                hasNextPage
                hasPreviousPage
                endCursor
                startCursor
            }
        }
    }
    ${savedSearchFragment}
`

export const savedSearchQuery = gql`
    query SavedSearch($id: ID!) {
        node(id: $id) {
            __typename
            ... on SavedSearch {
                ...SavedSearchFields
            }
        }
    }
    ${savedSearchFragment}
`

export const createSavedSearchMutation = gql`
    mutation CreateSavedSearch($input: SavedSearchInput!) {
        createSavedSearch(input: $input) {
            ...SavedSearchFields
        }
    }
    ${savedSearchFragment}
`

export const updateSavedSearchMutation = gql`
    mutation UpdateSavedSearch($id: ID!, $input: SavedSearchUpdateInput!) {
        updateSavedSearch(id: $id, input: $input) {
            ...SavedSearchFields
        }
    }
    ${savedSearchFragment}
`

export const transferSavedSearchOwnershipMutation = gql`
    mutation TransferSavedSearchOwnership($id: ID!, $newOwner: ID!) {
        transferSavedSearchOwnership(id: $id, newOwner: $newOwner) {
            ...SavedSearchFields
        }
    }
    ${savedSearchFragment}
`

export const changeSavedSearchVisibilityMutation = gql`
    mutation ChangeSavedSearchVisibility($id: ID!, $newVisibility: SavedSearchVisibility!) {
        changeSavedSearchVisibility(id: $id, newVisibility: $newVisibility) {
            ...SavedSearchFields
        }
    }
    ${savedSearchFragment}
`

export const deleteSavedSearchMutation = gql`
    mutation DeleteSavedSearch($id: ID!) {
        deleteSavedSearch(id: $id) {
            alwaysNil
        }
    }
`
