import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

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
