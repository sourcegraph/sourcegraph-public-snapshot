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
    BatchSpecApplyPreviewResult,
    AllChangesetSpecIDsVariables,
    AllChangesetSpecIDsResult,
    ChangesetSpecIDConnectionFields,
} from '../../../graphql-operations'

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

export const queryAllChangesetSpecIDs = ({
    batchSpec,
    search,
    currentState,
    action,
}: BatchSpecApplyPreviewVariables): Observable<Scalars['ID'][]> => {
    const request = (after: string | null): Observable<ChangesetSpecIDConnectionFields> =>
        requestGraphQL<AllChangesetSpecIDsResult, AllChangesetSpecIDsVariables>(
            gql`
                query AllChangesetSpecIDs(
                    $batchSpec: ID!
                    $after: String
                    $search: String
                    $currentState: ChangesetState
                    $action: ChangesetSpecOperation
                ) {
                    node(id: $batchSpec) {
                        __typename
                        ... on BatchSpec {
                            applyPreview(
                                first: 10000
                                after: $after
                                search: $search
                                currentState: $currentState
                                action: $action
                            ) {
                                ...ChangesetSpecIDConnectionFields
                            }
                        }
                    }
                }

                fragment ChangesetSpecIDConnectionFields on ChangesetApplyPreviewConnection {
                    nodes {
                        __typename
                        ... on VisibleChangesetApplyPreview {
                            targets {
                                ... on VisibleApplyPreviewTargetsAttach {
                                    changesetSpec {
                                        id
                                    }
                                }
                                ... on VisibleApplyPreviewTargetsUpdate {
                                    changesetSpec {
                                        id
                                    }
                                }
                            }
                        }
                    }
                    pageInfo {
                        hasNextPage
                        endCursor
                    }
                }
            `,
            { batchSpec, after, search, currentState, action }
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
        reduce(
            (previous, next) =>
                previous.concat(
                    next.nodes
                        .map(node => {
                            if (node.__typename !== 'VisibleChangesetApplyPreview') {
                                return undefined
                            }
                            return node.targets.changesetSpec.id
                        })
                        .filter((maybeID): maybeID is string => maybeID !== undefined)
                ),
            [] as Scalars['ID'][]
        )
    )
}
