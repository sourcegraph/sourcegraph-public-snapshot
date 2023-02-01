import { gql } from '@sourcegraph/http-client'

const listBatchChangeFragment = gql`
    fragment BatchChangesFields on BatchChangeConnection {
        nodes {
            __typename
            ...ListBatchChange
        }
        pageInfo {
            endCursor
            hasNextPage
        }
        totalCount
    }

    fragment ListBatchChange on BatchChange {
        id
        url
        name
        namespace {
            namespaceName
            url
        }
        description
        createdAt
        closedAt
        state
        changesetsStats {
            open
            closed
            merged
        }
        currentSpec {
            __typename
            ...ListBatchChangeLatestSpecFields
        }
        batchSpecs(first: 1) {
            nodes {
                __typename
                ...ListBatchChangeLatestSpecFields
            }
        }
    }

    fragment ListBatchChangeLatestSpecFields on BatchSpec {
        id
        state
        applyURL
    }
`

export const BATCH_CHANGES = gql`
    query BatchChanges($first: Int, $after: String, $states: [BatchChangeState!], $viewerCanAdminister: Boolean) {
        batchChanges(first: $first, after: $after, states: $states, viewerCanAdminister: $viewerCanAdminister) {
            __typename
            ...BatchChangesFields
        }
    }

    ${listBatchChangeFragment}
`
export const BATCH_CHANGES_BY_NAMESPACE = gql`
    query BatchChangesByNamespace(
        $namespaceID: ID!
        $first: Int
        $after: String
        $states: [BatchChangeState!]
        $viewerCanAdminister: Boolean
    ) {
        node(id: $namespaceID) {
            __typename
            ... on User {
                batchChanges(first: $first, after: $after, states: $states, viewerCanAdminister: $viewerCanAdminister) {
                    __typename
                    ...BatchChangesFields
                }
            }
            ... on Org {
                batchChanges(first: $first, after: $after, states: $states, viewerCanAdminister: $viewerCanAdminister) {
                    __typename
                    ...BatchChangesFields
                }
            }
        }
    }

    ${listBatchChangeFragment}
`

export const GET_LICENSE_AND_USAGE_INFO = gql`
    query GetLicenseAndUsageInfo {
        campaigns: enterpriseLicenseHasFeature(feature: "campaigns")
        batchChanges: enterpriseLicenseHasFeature(feature: "batch-changes")
        allBatchChanges: batchChanges(first: 1) {
            totalCount
        }
        maxUnlicensedChangesets
    }
`

export const GLOBAL_CHANGESETS_STATS = gql`
    query GlobalChangesetsStats {
        batchChanges {
            totalCount
        }
        globalChangesetsStats {
            open
            closed
            merged
        }
    }
`
