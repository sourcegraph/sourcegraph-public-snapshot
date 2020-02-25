import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { queryGraphQL, mutateGraphQL } from '../../../backend/graphql'
import { Observable } from 'rxjs'
import {
    ID,
    IEmptyResponse,
    IActionExecution,
    IActionsOnQueryArguments,
    IActionConnection,
} from '../../../../../shared/src/graphql/schema'
import { PreviewFileDiffFields, FileDiffHunkRangeFields, DiffStatFields } from '../../../backend/diff'

export async function retryActionJob(actionJobID: ID): Promise<IEmptyResponse | null> {
    const result = await mutateGraphQL(
        gql`
            mutation RetryActionJob($actionJob: ID!) {
                retryActionJob(actionJob: $actionJob) {
                    id
                }
            }
        `,
        { actionJob: actionJobID }
    ).toPromise()
    return dataOrThrowErrors(result).retryActionJob
}

export const fetchActionExecutionByID = (actionExecution: ID): Observable<IActionExecution | null> =>
    queryGraphQL(
        gql`
            query ActionExecutionByID($actionExecution: ID!) {
                node(id: $actionExecution) {
                    __typename
                    ... on ActionExecution {
                        id
                        action {
                            id
                            schedule
                            cancelPreviousScheduledExecution
                            savedSearch {
                                id
                                description
                            }
                            campaign {
                                id
                                name
                            }
                        }
                        definition {
                            steps
                            actionWorkspace {
                                name
                            }
                            env {
                                key
                            }
                        }
                        invokationReason
                        status {
                            errors
                            state
                            pendingCount
                            completedCount
                        }
                        campaignPlan {
                            id
                        }
                        jobs {
                            totalCount
                            nodes {
                                id
                                repository {
                                    name
                                }
                                baseRevision
                                state
                                runner {
                                    id
                                    name
                                    description
                                    state
                                }
                                executionStart
                                executionEnd
                                log
                                diff {
                                    fileDiffs {
                                        nodes {
                                            ...PreviewFileDiffFields
                                        }
                                        totalCount
                                        pageInfo {
                                            hasNextPage
                                        }
                                        diffStat {
                                            ...DiffStatFields
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }

            ${PreviewFileDiffFields}

            ${FileDiffHunkRangeFields}

            ${DiffStatFields}
        `,
        { actionExecution }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                return null
            }
            if (node.__typename !== 'ActionExecution') {
                throw new Error(`The given ID is a ${node.__typename}, not a ActionExecution`)
            }
            return node
        })
    )

export const queryActions = ({ first }: IActionsOnQueryArguments): Observable<IActionConnection> =>
    queryGraphQL(
        gql`
            query Actions($first: Int) {
                actions(first: $first) {
                    totalCount
                    nodes {
                        id
                        savedSearch {
                            description
                        }
                        schedule
                        actionExecutions {
                            totalCount
                            nodes {
                                id
                                definition {
                                    steps
                                    actionWorkspace {
                                        name
                                    }
                                    env {
                                        key
                                        value
                                    }
                                }
                                invokationReason
                                status {
                                    errors
                                    state
                                    pendingCount
                                    completedCount
                                }
                                campaignPlan {
                                    id
                                }
                            }
                        }
                    }
                }
            }
        `,
        { first }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.actions)
    )
