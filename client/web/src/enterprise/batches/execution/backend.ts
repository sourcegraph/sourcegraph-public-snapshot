import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../backend/graphql'
import {
    BatchSpecExecutionByIDResult,
    BatchSpecExecutionByIDVariables,
    BatchSpecExecutionFields,
    CancelBatchSpecExecutionResult,
    CancelBatchSpecExecutionVariables,
    Scalars,
} from '../../../graphql-operations'

const batchSpecExecutionFieldsFragment = gql`
    fragment BatchSpecExecutionFields on BatchSpec {
        id
        originalInput
        state
        createdAt
        startedAt
        finishedAt
        failureMessage
        applyURL
        creator {
            id
            url
            displayName
        }
        namespace {
            id
            url
            namespaceName
        }
        workspaceResolution {
            workspaces {
                totalCount
                pageInfo {
                    endCursor
                    hasNextPage
                }
                nodes {
                    id
                    steps {
                        run
                        diffStat {
                            added
                            changed
                            deleted
                        }
                        diff {
                            baseRepository {
                                name
                            }
                        }
                        container
                        cachedResultFound
                        skipped
                        outputLines
                        startedAt
                        finishedAt
                        exitCode
                        environment {
                            name
                            value
                        }
                        outputVariables {
                            name
                            value
                        }
                    }
                    searchResultPaths
                    startedAt
                    finishedAt
                    failureMessage
                    state
                    changesetSpecs {
                        id
                        type
                        __typename
                        ... on VisibleChangesetSpec {
                            description {
                                ... on GitBranchChangesetDescription {
                                    title
                                }
                                ... on ExistingChangesetReference {
                                    externalID
                                }
                            }
                        }
                    }
                    placeInQueue
                    repository {
                        name
                        url
                    }
                    branch {
                        name
                    }
                    path
                    onlyFetchWorkspace
                    ignored
                    unsupported
                    cachedResultFound
                    stages {
                        setup {
                            key
                            command
                            startTime
                            exitCode
                            out
                            durationMilliseconds
                        }
                        srcExec {
                            key
                            command
                            startTime
                            exitCode
                            out
                            durationMilliseconds
                        }
                        teardown {
                            key
                            command
                            startTime
                            exitCode
                            out
                            durationMilliseconds
                        }
                    }
                }
            }
        }
    }
`

export const fetchBatchSpecExecution = (id: Scalars['ID']): Observable<BatchSpecExecutionFields | null> =>
    requestGraphQL<BatchSpecExecutionByIDResult, BatchSpecExecutionByIDVariables>(
        gql`
            query BatchSpecExecutionByID($id: ID!) {
                node(id: $id) {
                    __typename
                    ...BatchSpecExecutionFields
                }
            }
            ${batchSpecExecutionFieldsFragment}
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                return null
            }
            if (node.__typename !== 'BatchSpec') {
                throw new Error(`Node is a ${node.__typename}, not a BatchSpec`)
            }
            return node
        })
    )

export async function cancelBatchSpecExecution(id: Scalars['ID']): Promise<BatchSpecExecutionFields> {
    const result = await requestGraphQL<CancelBatchSpecExecutionResult, CancelBatchSpecExecutionVariables>(
        gql`
            mutation CancelBatchSpecExecution($id: ID!) {
                cancelBatchSpecExecution(batchSpec: $id) {
                    ...BatchSpecExecutionFields
                }
            }

            ${batchSpecExecutionFieldsFragment}
        `,
        { id }
    ).toPromise()
    return dataOrThrowErrors(result).cancelBatchSpecExecution
}
