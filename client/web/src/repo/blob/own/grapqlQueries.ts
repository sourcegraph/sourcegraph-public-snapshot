import { gql } from '@sourcegraph/http-client'

import { gitCommitFragment } from '../../commits/RepositoryCommitsPage'

export const OWNER_FIELDS = gql`
    fragment OwnerFields on Owner {
        __typename
        ... on Person {
            displayName
            email
            avatarURL
            user {
                id
                username
                displayName
                url
                primaryEmail {
                    email
                }
            }
        }
        ... on Team {
            id
            name
            teamDisplayName: displayName
            avatarURL
            url
            external
        }
    }
`

export const RECENT_CONTRIBUTOR_FIELDS = gql`
    fragment RecentContributorOwnershipSignalFields on RecentContributorOwnershipSignal {
        title
        description
    }
`

export const RECENT_VIEW_FIELDS = gql`
    fragment RecentViewOwnershipSignalFields on RecentViewOwnershipSignal {
        title
        description
    }
`

export const ASSIGNED_OWNER_FIELDS = gql`
    fragment AssignedOwnerFields on AssignedOwner {
        title
        description
        isDirectMatch
    }
`

export const FETCH_OWNERS = gql`
    ${OWNER_FIELDS}
    ${RECENT_CONTRIBUTOR_FIELDS}
    ${RECENT_VIEW_FIELDS}
    ${ASSIGNED_OWNER_FIELDS}

    fragment CodeownersFileEntryFields on CodeownersFileEntry {
        title
        description
        codeownersFile {
            url
        }
        ruleLineMatch
    }

    fragment BlobOwnershipFields on GitCommit {
        blob(path: $currentPath) {
            ownership {
                totalOwners
                nodes {
                    owner {
                        ...OwnerFields
                    }
                    reasons {
                        ...CodeownersFileEntryFields
                        ...RecentContributorOwnershipSignalFields
                        ...RecentViewOwnershipSignalFields
                        ...AssignedOwnerFields
                    }
                }
            }
        }
    }

    query FetchOwnership($repo: ID!, $revision: String!, $currentPath: String!) {
        node(id: $repo) {
            ... on Repository {
                commit(rev: $revision) {
                    ...BlobOwnershipFields
                }
                changelist(cid: $revision) {
                    commit {
                        ...BlobOwnershipFields
                    }
                }
            }
        }
        currentUser {
            permissions {
                nodes {
                    displayName
                }
            }
        }
    }
`

export const FETCH_TREE_OWNERS = gql`
    ${OWNER_FIELDS}
    ${RECENT_CONTRIBUTOR_FIELDS}
    ${RECENT_VIEW_FIELDS}
    ${ASSIGNED_OWNER_FIELDS}

    fragment CodeownersFileEntryFields on CodeownersFileEntry {
        title
        description
        codeownersFile {
            url
        }
        ruleLineMatch
    }

    fragment OwnershipConnectionFields on OwnershipConnection {
        totalOwners
        nodes {
            owner {
                ...OwnerFields
            }
            reasons {
                ...CodeownersFileEntryFields
                ...RecentContributorOwnershipSignalFields
                ...RecentViewOwnershipSignalFields
                ...AssignedOwnerFields
            }
        }
    }

    query FetchTreeOwnership($repo: ID!, $revision: String!, $currentPath: String!) {
        node(id: $repo) {
            ... on Repository {
                commit(rev: $revision) {
                    tree(path: $currentPath) {
                        ownership {
                            ...OwnershipConnectionFields
                        }
                    }
                }
            }
        }
        currentUser {
            permissions {
                nodes {
                    displayName
                }
            }
        }
    }
`

export const FETCH_OWNERS_AND_HISTORY = gql`
    ${OWNER_FIELDS}
    ${gitCommitFragment}

    fragment BlobOwnership on GitBlob {
        ownership(first: 2, reasons: [CODEOWNERS_FILE_ENTRY, ASSIGNED_OWNER]) {
            nodes {
                owner {
                    ...OwnerFields
                }
            }
            totalCount
        }
        contributors: ownership(reasons: [RECENT_CONTRIBUTOR_OWNERSHIP_SIGNAL]) {
            totalCount
        }
    }

    fragment HistoryFragment on GitCommit {
        ancestors(first: 1, path: $currentPath) {
            nodes {
                ...GitCommitFields
            }
        }
    }

    query FetchOwnersAndHistory($repo: ID!, $revision: String!, $currentPath: String!, $includeOwn: Boolean!) {
        node(id: $repo) {
            ... on Repository {
                __typename
                id
                sourceType
                commit(rev: $revision) {
                    __typename
                    blob(path: $currentPath) @include(if: $includeOwn) {
                        ...BlobOwnership
                    }
                    ...HistoryFragment
                }
                changelist(cid: $revision) {
                    __typename
                    commit {
                        blob(path: $currentPath) @include(if: $includeOwn) {
                            ...BlobOwnership
                        }
                        ...HistoryFragment
                    }
                }
            }
        }
    }
`

export const ASSIGN_OWNER = gql`
    mutation AssignOwner($input: AssignOwnerOrTeamInput!) {
        assignOwner(input: $input) {
            alwaysNil
        }
    }
`

export const REMOVE_ASSIGNED_OWNER = gql`
    mutation RemoveAssignedOwner($input: AssignOwnerOrTeamInput!) {
        removeAssignedOwner(input: $input) {
            alwaysNil
        }
    }
`

export const REMOVE_ASSIGNED_TEAM = gql`
    mutation RemoveAssignedTeam($input: AssignOwnerOrTeamInput!) {
        removeAssignedTeam(input: $input) {
            alwaysNil
        }
    }
`
