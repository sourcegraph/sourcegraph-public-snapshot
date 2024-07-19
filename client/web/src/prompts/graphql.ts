import { gql } from '@sourcegraph/http-client'

const promptFragment = gql`
    fragment PromptFields on Prompt {
        __typename
        id
        name
        description
        definition {
            text
        }
        draft
        owner {
            __typename
            id
            namespaceName
            ... on Org {
                displayName
            }
        }
        visibility
        nameWithOwner
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

export const promptsQuery = gql`
    query Prompts(
        $query: String
        $owner: ID = null
        $viewerIsAffiliated: Boolean
        $includeDrafts: Boolean = true
        $first: Int
        $last: Int
        $after: String
        $before: String
        $orderBy: PromptsOrderBy!
    ) {
        prompts(
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
                ...PromptFields
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
    ${promptFragment}
`

export const promptQuery = gql`
    query Prompt($id: ID!) {
        node(id: $id) {
            __typename
            ... on Prompt {
                ...PromptFields
            }
        }
    }
    ${promptFragment}
`

export const createPromptMutation = gql`
    mutation CreatePrompt($input: PromptInput!) {
        createPrompt(input: $input) {
            ...PromptFields
        }
    }
    ${promptFragment}
`

export const updatePromptMutation = gql`
    mutation UpdatePrompt($id: ID!, $input: PromptUpdateInput!) {
        updatePrompt(id: $id, input: $input) {
            ...PromptFields
        }
    }
    ${promptFragment}
`

export const transferPromptOwnershipMutation = gql`
    mutation TransferPromptOwnership($id: ID!, $newOwner: ID!) {
        transferPromptOwnership(id: $id, newOwner: $newOwner) {
            ...PromptFields
        }
    }
    ${promptFragment}
`

export const changePromptVisibilityMutation = gql`
    mutation ChangePromptVisibility($id: ID!, $newVisibility: PromptVisibility!) {
        changePromptVisibility(id: $id, newVisibility: $newVisibility) {
            ...PromptFields
        }
    }
    ${promptFragment}
`

export const deletePromptMutation = gql`
    mutation DeletePrompt($id: ID!) {
        deletePrompt(id: $id) {
            alwaysNil
        }
    }
`
