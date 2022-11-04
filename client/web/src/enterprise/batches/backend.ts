import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../backend/graphql'
import {
    BatchSpecsVariables,
    BatchSpecsResult,
    Scalars,
    BatchChangeBatchSpecsVariables,
    BatchChangeBatchSpecsResult,
    BatchSpecListConnectionFields,
} from '../../graphql-operations'

export const queryBatchSpecs = ({
    first,
    after,
    includeLocallyExecutedSpecs,
}: BatchSpecsVariables): Observable<BatchSpecListConnectionFields> =>
    requestGraphQL<BatchSpecsResult, BatchSpecsVariables>(
        gql`
            query BatchSpecs($first: Int, $after: String, $includeLocallyExecutedSpecs: Boolean) {
                batchSpecs(first: $first, after: $after, includeLocallyExecutedSpecs: $includeLocallyExecutedSpecs) {
                    ...BatchSpecListConnectionFields
                }
            }

            ${BATCH_SPEC_LIST_CONNECTION_FIELDS}
        `,
        {
            first,
            after,
            includeLocallyExecutedSpecs,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.batchSpecs)
    )

export const queryBatchChangeBatchSpecs = (id: Scalars['ID']) => ({
    first,
    after,
    includeLocallyExecutedSpecs,
}: Omit<BatchChangeBatchSpecsVariables, 'id'>): Observable<BatchSpecListConnectionFields> =>
    requestGraphQL<BatchChangeBatchSpecsResult, BatchChangeBatchSpecsVariables>(
        gql`
            query BatchChangeBatchSpecs($id: ID!, $first: Int, $after: String, $includeLocallyExecutedSpecs: Boolean) {
                node(id: $id) {
                    __typename
                    ... on BatchChange {
                        batchSpecs(
                            first: $first
                            after: $after
                            includeLocallyExecutedSpecs: $includeLocallyExecutedSpecs
                        ) {
                            ...BatchSpecListConnectionFields
                        }
                    }
                }
            }

            ${BATCH_SPEC_LIST_CONNECTION_FIELDS}
        `,
        {
            id,
            first,
            after,
            includeLocallyExecutedSpecs,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.node) {
                throw new Error('Batch change not found')
            }
            if (data.node.__typename !== 'BatchChange') {
                throw new Error(`Node is a ${data.node.__typename}, not a BatchChange`)
            }
            return data.node.batchSpecs
        })
    )

const BATCH_SPEC_LIST_FIELDS_FRAGMENT = gql`
    fragment BatchSpecListFields on BatchSpec {
        __typename
        id
        state
        startedAt
        finishedAt
        createdAt
        source
        description {
            name
        }
        namespace {
            namespaceName
            url
        }
        creator {
            username
        }
        originalInput
    }
`

const BATCH_SPEC_LIST_CONNECTION_FIELDS = gql`
    fragment BatchSpecListConnectionFields on BatchSpecConnection {
        totalCount
        pageInfo {
            endCursor
            hasNextPage
        }
        nodes {
            ...BatchSpecListFields
        }
    }

    ${BATCH_SPEC_LIST_FIELDS_FRAGMENT}
`
