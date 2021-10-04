import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../../backend/graphql'
import {
    BatchSpecByID2Result,
    BatchSpecByID2Variables,
    BatchSpecWorkspacesFields,
    CreateBatchSpecFromRawResult,
    CreateBatchSpecFromRawVariables,
    ReplaceBatchSpecInputResult,
    ReplaceBatchSpecInputVariables,
    Scalars,
} from '../../../../graphql-operations'

const fragment = gql`
    fragment BatchSpecWorkspacesFields on BatchSpec {
        id
        originalInput
        workspaceResolution {
            workspaces {
                nodes {
                    ...BatchSpecWorkspaceFields
                }
            }
            state
            failureMessage
        }
        allowUnsupported
        allowIgnored
    }

    fragment BatchSpecWorkspaceFields on BatchSpecWorkspace {
        repository {
            id
            name
            url
        }
        ignored
        unsupported
        branch {
            id
            abbrevName
            displayName
            target {
                oid
            }
        }
        path
        onlyFetchWorkspace
        steps {
            run
            container
        }
        searchResultPaths
    }
`

export function createBatchSpecFromRaw(spec: string): Observable<BatchSpecWorkspacesFields> {
    return requestGraphQL<CreateBatchSpecFromRawResult, CreateBatchSpecFromRawVariables>(
        gql`
            mutation CreateBatchSpecFromRaw($spec: String!) {
                createBatchSpecFromRaw(batchSpec: $spec) {
                    ...BatchSpecWorkspacesFields
                }
            }

            ${fragment}
        `,
        { spec }
    ).pipe(
        map(dataOrThrowErrors),
        map(result => result.createBatchSpecFromRaw)
    )
}

export function replaceBatchSpecInput(
    previousSpec: Scalars['ID'],
    spec: string
): Observable<BatchSpecWorkspacesFields> {
    return requestGraphQL<ReplaceBatchSpecInputResult, ReplaceBatchSpecInputVariables>(
        gql`
            mutation ReplaceBatchSpecInput($previousSpec: ID!, $spec: String!) {
                replaceBatchSpecInput(previousSpec: $previousSpec, batchSpec: $spec) {
                    ...BatchSpecWorkspacesFields
                }
            }

            ${fragment}
        `,
        { previousSpec, spec }
    ).pipe(
        map(dataOrThrowErrors),
        map(result => result.replaceBatchSpecInput)
    )
}

export function fetchBatchSpec(id: Scalars['ID']): Observable<BatchSpecWorkspacesFields> {
    return requestGraphQL<BatchSpecByID2Result, BatchSpecByID2Variables>(
        gql`
            query BatchSpecByID2($id: ID!) {
                node(id: $id) {
                    __typename
                    ...BatchSpecWorkspacesFields
                }
            }

            ${fragment}
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.node) {
                throw new Error('Not found')
            }
            if (data.node.__typename !== 'BatchSpec') {
                throw new Error(`Node is a ${data.node.__typename}, not a BatchSpec`)
            }
            return data.node
        })
    )
}
