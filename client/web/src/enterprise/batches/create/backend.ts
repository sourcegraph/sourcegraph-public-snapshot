import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../backend/graphql'
import {
    ExecuteBatchSpecFields,
    ExecuteBatchSpecResult,
    ExecuteBatchSpecVariables,
    BatchSpecWorkspacesByIDResult,
    BatchSpecWorkspacesByIDVariables,
    BatchSpecWithWorkspacesFields,
    CreateBatchSpecFromRawResult,
    CreateBatchSpecFromRawVariables,
    ReplaceBatchSpecInputResult,
    ReplaceBatchSpecInputVariables,
    Scalars,
} from '../../../graphql-operations'

export async function executeBatchSpec(spec: Scalars['ID']): Promise<ExecuteBatchSpecFields> {
    const result = await requestGraphQL<ExecuteBatchSpecResult, ExecuteBatchSpecVariables>(
        gql`
            mutation ExecuteBatchSpec($id: ID!) {
                executeBatchSpec(batchSpec: $id) {
                    ...ExecuteBatchSpecFields
                }
            }

            fragment ExecuteBatchSpecFields on BatchSpec {
                id
                namespace {
                    url
                }
            }
        `,
        { id: spec }
    ).toPromise()
    return dataOrThrowErrors(result).executeBatchSpec
}

const fragment = gql`
    fragment BatchSpecWithWorkspacesFields on BatchSpec {
        id
        originalInput
        workspaceResolution {
            workspaces(first: 10000) {
                nodes {
                    ...CreateBatchSpecWorkspaceFields
                }
            }
            state
            failureMessage
        }
        allowUnsupported
        allowIgnored
        importingChangesets(first: 10000) {
            totalCount
            nodes {
                __typename
                id
                ... on VisibleChangesetSpec {
                    description {
                        __typename
                        ... on ExistingChangesetReference {
                            baseRepository {
                                name
                                url
                            }
                            externalID
                        }
                    }
                }
            }
        }
    }

    fragment CreateBatchSpecWorkspaceFields on BatchSpecWorkspace {
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
        cachedResultFound
    }
`

export function createBatchSpecFromRaw(
    spec: string,
    namespace: Scalars['ID']
): Observable<BatchSpecWithWorkspacesFields> {
    return requestGraphQL<CreateBatchSpecFromRawResult, CreateBatchSpecFromRawVariables>(
        gql`
            mutation CreateBatchSpecFromRaw($spec: String!, $namespace: ID!) {
                createBatchSpecFromRaw(batchSpec: $spec, namespace: $namespace) {
                    ...BatchSpecWithWorkspacesFields
                }
            }

            ${fragment}
        `,
        { spec, namespace }
    ).pipe(
        map(dataOrThrowErrors),
        map(result => result.createBatchSpecFromRaw)
    )
}

export function replaceBatchSpecInput(
    previousSpec: Scalars['ID'],
    spec: string
): Observable<BatchSpecWithWorkspacesFields> {
    return requestGraphQL<ReplaceBatchSpecInputResult, ReplaceBatchSpecInputVariables>(
        gql`
            mutation ReplaceBatchSpecInput($previousSpec: ID!, $spec: String!) {
                replaceBatchSpecInput(previousSpec: $previousSpec, batchSpec: $spec) {
                    ...BatchSpecWithWorkspacesFields
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

export function fetchBatchSpec(id: Scalars['ID']): Observable<BatchSpecWithWorkspacesFields> {
    return requestGraphQL<BatchSpecWorkspacesByIDResult, BatchSpecWorkspacesByIDVariables>(
        gql`
            query BatchSpecWorkspacesByID($id: ID!) {
                node(id: $id) {
                    __typename
                    ...BatchSpecWithWorkspacesFields
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
