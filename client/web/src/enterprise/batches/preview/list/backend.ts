import { EMPTY, type Observable } from 'rxjs'
import { expand, map, reduce } from 'rxjs/operators'

import { gql, dataOrThrowErrors } from '@sourcegraph/http-client'

import { diffStatFields, fileDiffFields } from '../../../../backend/diff'
import { requestGraphQL } from '../../../../backend/graphql'
import type {
    ChangesetSpecFileDiffsVariables,
    ChangesetSpecFileDiffsResult,
    BatchSpecApplyPreviewConnectionFields,
    BatchSpecApplyPreviewResult,
    BatchSpecApplyPreviewVariables,
    ChangesetSpecFileDiffConnectionFields,
    Scalars,
    AllPublishableChangesetSpecIDsVariables,
    PublishableChangesetSpecIDsConnectionFields,
    AllPublishableChangesetSpecIDsResult,
} from '../../../../graphql-operations'
import { personLinkFieldsFragment } from '../../../../person/PersonLink'
import { filterPublishableIDs } from '../utils'

const changesetSpecFieldsFragment = gql`
    fragment CommonChangesetSpecFields on ChangesetSpec {
        type
    }

    fragment HiddenChangesetSpecFields on HiddenChangesetSpec {
        __typename
        id
        ...CommonChangesetSpecFields
    }

    fragment VisibleChangesetSpecFields on VisibleChangesetSpec {
        __typename
        id
        ...CommonChangesetSpecFields
        description {
            __typename
            ...ExistingChangesetReferenceFields
            ...GitBranchChangesetDescriptionFields
        }
        forkTarget {
            pushUser
            namespace
        }
    }

    fragment ExistingChangesetReferenceFields on ExistingChangesetReference {
        baseRepository {
            name
            url
        }
        externalID
    }

    fragment GitBranchChangesetDescriptionFields on GitBranchChangesetDescription {
        baseRepository {
            name
            url
        }
        title
        published
        body
        commits {
            author {
                avatarURL
                email
                displayName
                user {
                    username
                    displayName
                    url
                }
            }
            subject
            body
        }
        baseRef
        headRef
        diffStat {
            ...DiffStatFields
        }
    }

    ${diffStatFields}
`

const batchSpecApplyPreviewConnectionFieldsFragment = gql`
    fragment BatchSpecApplyPreviewConnectionFields on ChangesetApplyPreviewConnection {
        totalCount
        pageInfo {
            endCursor
            hasNextPage
        }
        nodes {
            ...ChangesetApplyPreviewFields
        }
    }

    fragment ChangesetApplyPreviewFields on ChangesetApplyPreview {
        __typename
        ... on HiddenChangesetApplyPreview {
            ...HiddenChangesetApplyPreviewFields
        }
        ... on VisibleChangesetApplyPreview {
            ...VisibleChangesetApplyPreviewFields
        }
    }

    fragment HiddenChangesetApplyPreviewFields on HiddenChangesetApplyPreview {
        __typename
        targets {
            __typename
            ... on HiddenApplyPreviewTargetsAttach {
                changesetSpec {
                    ...HiddenChangesetSpecFields
                }
            }
            ... on HiddenApplyPreviewTargetsUpdate {
                changesetSpec {
                    ...HiddenChangesetSpecFields
                }
                changeset {
                    id
                    state
                }
            }
            ... on HiddenApplyPreviewTargetsDetach {
                changeset {
                    id
                    state
                }
            }
        }
    }

    fragment VisibleChangesetApplyPreviewFields on VisibleChangesetApplyPreview {
        __typename
        operations
        delta {
            titleChanged
            bodyChanged
            baseRefChanged
            diffChanged
            authorEmailChanged
            authorNameChanged
            commitMessageChanged
        }
        targets {
            __typename
            ... on VisibleApplyPreviewTargetsAttach {
                changesetSpec {
                    ...VisibleChangesetSpecFields
                }
            }
            ... on VisibleApplyPreviewTargetsUpdate {
                changesetSpec {
                    ...VisibleChangesetSpecFields
                }
                changeset {
                    id
                    title
                    state
                    externalURL {
                        url
                    }
                    externalID
                    currentSpec {
                        description {
                            __typename
                            ... on GitBranchChangesetDescription {
                                baseRef
                                title
                                body
                                commits {
                                    author {
                                        avatarURL
                                        email
                                        displayName
                                        user {
                                            username
                                            displayName
                                            url
                                        }
                                    }
                                    body
                                    subject
                                }
                            }
                        }
                    }
                    author {
                        ...PersonLinkFields
                    }
                }
            }
            ... on VisibleApplyPreviewTargetsDetach {
                changeset {
                    id
                    title
                    state
                    externalURL {
                        url
                    }
                    externalID
                    repository {
                        url
                        name
                    }
                    diffStat {
                        added
                        deleted
                    }
                }
            }
        }
    }

    ${changesetSpecFieldsFragment}

    ${personLinkFieldsFragment}
`

export const queryChangesetApplyPreview = ({
    batchSpec,
    first,
    after,
    search,
    currentState,
    action,
    publicationStates,
}: BatchSpecApplyPreviewVariables): Observable<BatchSpecApplyPreviewConnectionFields> =>
    requestGraphQL<BatchSpecApplyPreviewResult, BatchSpecApplyPreviewVariables>(
        gql`
            query BatchSpecApplyPreview(
                $batchSpec: ID!
                $first: Int
                $after: String
                $search: String
                $currentState: ChangesetState
                $action: ChangesetSpecOperation
                $publicationStates: [ChangesetSpecPublicationStateInput!]
            ) {
                node(id: $batchSpec) {
                    __typename
                    ... on BatchSpec {
                        applyPreview(
                            first: $first
                            after: $after
                            search: $search
                            currentState: $currentState
                            action: $action
                            publicationStates: $publicationStates
                        ) {
                            ...BatchSpecApplyPreviewConnectionFields
                        }
                    }
                }
            }

            ${batchSpecApplyPreviewConnectionFieldsFragment}
        `,
        { batchSpec, first, after, search, currentState, action, publicationStates }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error(`BatchSpec with ID ${batchSpec} does not exist`)
            }
            if (node.__typename !== 'BatchSpec') {
                throw new Error(`The given ID is a ${node.__typename}, not a BatchSpec`)
            }
            return node.applyPreview
        })
    )

const changesetSpecFileDiffsFields = gql`
    fragment ChangesetSpecFileDiffsFields on VisibleChangesetSpec {
        description {
            __typename
            ... on GitBranchChangesetDescription {
                diff {
                    fileDiffs(first: $first, after: $after) {
                        ...ChangesetSpecFileDiffConnectionFields
                    }
                }
            }
        }
    }

    fragment ChangesetSpecFileDiffConnectionFields on FileDiffConnection {
        nodes {
            ...FileDiffFields
        }
        totalCount
        pageInfo {
            hasNextPage
            endCursor
        }
    }

    ${fileDiffFields}
`

export const queryChangesetSpecFileDiffs = ({
    changesetSpec,
    first,
    after,
}: ChangesetSpecFileDiffsVariables): Observable<ChangesetSpecFileDiffConnectionFields> =>
    requestGraphQL<ChangesetSpecFileDiffsResult, ChangesetSpecFileDiffsVariables>(
        gql`
            query ChangesetSpecFileDiffs($changesetSpec: ID!, $first: Int, $after: String) {
                node(id: $changesetSpec) {
                    __typename
                    ...ChangesetSpecFileDiffsFields
                }
            }

            ${changesetSpecFileDiffsFields}
        `,
        { changesetSpec, first, after }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error(`ChangesetSpec with ID ${changesetSpec} does not exist`)
            }
            if (node.__typename !== 'VisibleChangesetSpec') {
                throw new Error(`The given ID is a ${node.__typename}, not a VisibleChangesetSpec`)
            }
            if (node.description.__typename !== 'GitBranchChangesetDescription') {
                throw new Error('The given ChangesetSpec is no GitBranchChangesetDescription')
            }
            return node.description.diff.fileDiffs
        })
    )

const publishableChangesetSpecIDsFieldsFragment = gql`
    fragment PublishableChangesetSpecIDsConnectionFields on ChangesetApplyPreviewConnection {
        nodes {
            ...PublishableChangesetSpecIDsChangesetApplyPreviewFields
        }
        pageInfo {
            hasNextPage
            endCursor
        }
    }

    fragment PublishableChangesetSpecIDsChangesetApplyPreviewFields on ChangesetApplyPreview {
        __typename
        ... on HiddenChangesetApplyPreview {
            ...PublishableChangesetSpecIDsHiddenChangesetApplyPreviewFields
        }
        ... on VisibleChangesetApplyPreview {
            ...PublishableChangesetSpecIDsVisibleChangesetApplyPreviewFields
        }
    }

    fragment PublishableChangesetSpecIDsHiddenChangesetApplyPreviewFields on HiddenChangesetApplyPreview {
        __typename
        targets {
            __typename
        }
    }

    fragment PublishableChangesetSpecIDsVisibleChangesetApplyPreviewFields on VisibleChangesetApplyPreview {
        __typename
        targets {
            __typename
            ... on VisibleApplyPreviewTargetsAttach {
                changesetSpec {
                    ...PublishableChangesetSpecIDsVisibleChangesetSpecFields
                }
            }
            ... on VisibleApplyPreviewTargetsUpdate {
                changesetSpec {
                    ...PublishableChangesetSpecIDsVisibleChangesetSpecFields
                }
                changeset {
                    state
                }
            }
        }
    }

    fragment PublishableChangesetSpecIDsVisibleChangesetSpecFields on VisibleChangesetSpec {
        id
        description {
            __typename
            ... on GitBranchChangesetDescription {
                published
            }
        }
    }
`

export const queryPublishableChangesetSpecIDs = ({
    batchSpec,
    first,
    search,
    currentState,
    action,
}: Omit<AllPublishableChangesetSpecIDsVariables, 'after'>): Observable<Scalars['ID'][]> => {
    const request = (after: string | null): Observable<PublishableChangesetSpecIDsConnectionFields> =>
        requestGraphQL<AllPublishableChangesetSpecIDsResult, AllPublishableChangesetSpecIDsVariables>(
            gql`
                query AllPublishableChangesetSpecIDs(
                    $batchSpec: ID!
                    $first: Int
                    $after: String
                    $search: String
                    $currentState: ChangesetState
                    $action: ChangesetSpecOperation
                ) {
                    node(id: $batchSpec) {
                        __typename
                        ... on BatchSpec {
                            applyPreview(
                                first: $first
                                after: $after
                                search: $search
                                currentState: $currentState
                                action: $action
                            ) {
                                ...PublishableChangesetSpecIDsConnectionFields
                            }
                        }
                    }
                }

                ${publishableChangesetSpecIDsFieldsFragment}
            `,
            { batchSpec, first, after, search, currentState, action }
        ).pipe(
            map(dataOrThrowErrors),
            map(({ node }) => {
                if (!node) {
                    throw new Error(`BatchSpec with ID ${batchSpec} does not exist`)
                }
                if (node.__typename !== 'BatchSpec') {
                    throw new Error(`The given ID is a ${node.__typename}, not a BatchSpec`)
                }
                return node.applyPreview
            })
        )

    return request(null).pipe(
        expand(connection => (connection.pageInfo.hasNextPage ? request(connection.pageInfo.endCursor) : EMPTY)),
        reduce<PublishableChangesetSpecIDsConnectionFields, Scalars['ID'][]>(
            (previous, next) => previous.concat(filterPublishableIDs(next.nodes)),
            []
        )
    )
}
