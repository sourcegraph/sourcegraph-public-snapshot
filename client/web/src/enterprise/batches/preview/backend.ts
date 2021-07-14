import { EMPTY, Observable } from 'rxjs'
import { expand, map, reduce } from 'rxjs/operators'

import { gql, dataOrThrowErrors } from '@sourcegraph/shared/src/graphql/graphql'

import { diffStatFields } from '../../../backend/diff'
import { requestGraphQL } from '../../../backend/graphql'
import {
    Scalars,
    CreateBatchChangeVariables,
    CreateBatchChangeResult,
    ApplyBatchChangeResult,
    ApplyBatchChangeVariables,
    BatchSpecByIDResult,
    BatchSpecByIDVariables,
    BatchSpecFields,
    BatchSpecApplyPreviewVariables,
    BatchSpecApplyPreviewConnectionFields,
    BatchSpecApplyPreviewResult,
} from '../../../graphql-operations'
import { personLinkFieldsFragment } from '../../../person/PersonLink'

import { canSetPublishedState } from './utils'

export const viewerBatchChangesCodeHostsFragment = gql`
    fragment ViewerBatchChangesCodeHostsFields on BatchChangesCodeHostConnection {
        totalCount
        nodes {
            externalServiceURL
            externalServiceKind
        }
    }
`

const supersedingBatchSpecFragment = gql`
    fragment SupersedingBatchSpecFields on BatchSpec {
        createdAt
        applyURL
    }
`

export const batchSpecFragment = gql`
    fragment BatchSpecFields on BatchSpec {
        id
        description {
            name
            description
        }
        appliesToBatchChange {
            id
            name
            url
        }
        createdAt
        creator {
            username
            url
        }
        expiresAt
        namespace {
            namespaceName
            url
        }
        viewerCanAdminister
        diffStat {
            ...DiffStatFields
        }
        originalInput
        applyPreview {
            stats {
                close
                detach
                archive
                import
                publish
                publishDraft
                push
                reopen
                undraft
                update

                added
                modified
                removed

                uiPublished
            }
            totalCount
        }
        supersedingBatchSpec {
            ...SupersedingBatchSpecFields
        }
        viewerBatchChangesCodeHosts(onlyWithoutCredential: true) {
            ...ViewerBatchChangesCodeHostsFields
        }
    }

    ${viewerBatchChangesCodeHostsFragment}

    ${diffStatFields}

    ${supersedingBatchSpecFragment}
`

export const fetchBatchSpecById = (batchSpec: Scalars['ID']): Observable<BatchSpecFields | null> =>
    requestGraphQL<BatchSpecByIDResult, BatchSpecByIDVariables>(
        gql`
            query BatchSpecByID($batchSpec: ID!) {
                node(id: $batchSpec) {
                    __typename
                    ... on BatchSpec {
                        ...BatchSpecFields
                    }
                }
            }
            ${batchSpecFragment}
        `,
        { batchSpec }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                return null
            }
            if (node.__typename !== 'BatchSpec') {
                throw new Error(`The given ID is a ${node.__typename}, not a BatchSpec`)
            }
            return node
        })
    )

export const createBatchChange = ({
    batchSpec,
    publicationStates,
}: CreateBatchChangeVariables): Promise<CreateBatchChangeResult['createBatchChange']> =>
    requestGraphQL<CreateBatchChangeResult, CreateBatchChangeVariables>(
        gql`
            mutation CreateBatchChange($batchSpec: ID!, $publicationStates: [ChangesetSpecPublicationStateInput!]) {
                createBatchChange(batchSpec: $batchSpec, publicationStates: $publicationStates) {
                    id
                    url
                }
            }
        `,
        { batchSpec, publicationStates }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.createBatchChange)
        )
        .toPromise()

export const applyBatchChange = ({
    batchSpec,
    batchChange,
    publicationStates,
}: ApplyBatchChangeVariables): Promise<ApplyBatchChangeResult['applyBatchChange']> =>
    requestGraphQL<ApplyBatchChangeResult, ApplyBatchChangeVariables>(
        gql`
            mutation ApplyBatchChange(
                $batchSpec: ID!
                $batchChange: ID!
                $publicationStates: [ChangesetSpecPublicationStateInput!]
            ) {
                applyBatchChange(
                    batchSpec: $batchSpec
                    ensureBatchChange: $batchChange
                    publicationStates: $publicationStates
                ) {
                    id
                    url
                }
            }
        `,
        { batchSpec, batchChange, publicationStates }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.applyBatchChange)
        )
        .toPromise()

export const queryChangesetApplyPreview = ({
    batchSpec,
    first,
    after,
    search,
    currentState,
    action,
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
                            ...BatchSpecApplyPreviewConnectionFields
                        }
                    }
                }
            }

            ${batchSpecApplyPreviewConnectionFieldsFragment}
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

export const queryPublishableChangesetSpecs = ({
    batchSpec,
    search,
    currentState,
    action,
}: BatchSpecApplyPreviewVariables): Observable<Scalars['ID'][]> => {
    const request = (after: string | null): Observable<BatchSpecApplyPreviewConnectionFields> =>
        queryChangesetApplyPreview({
            batchSpec,
            first: 10000,
            after,
            search,
            currentState,
            action,
        })

    return request(null).pipe(
        expand(connection => (connection.pageInfo.hasNextPage ? request(connection.pageInfo.endCursor) : EMPTY)),
        reduce(
            (previous, next) =>
                previous.concat(
                    next.nodes
                        .map(node => canSetPublishedState(node))
                        .filter((maybeID): maybeID is string => maybeID !== null)
                ),
            [] as Scalars['ID'][]
        )
    )
}

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

export const batchSpecApplyPreviewConnectionFieldsFragment = gql`
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
                        changed
                        deleted
                    }
                }
            }
        }
    }

    ${changesetSpecFieldsFragment}

    ${personLinkFieldsFragment}
`
