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
    IAction,
    ActionJobState,
    IActionJob,
} from '../../../../../shared/src/graphql/schema'
import { PreviewFileDiffFields, FileDiffHunkRangeFields, DiffStatFields } from '../../../backend/diff'

export async function retryActionJob(actionJobID: ID): Promise<IEmptyResponse | null> {
    const result = await mutateGraphQL(
        gql`
            mutation RetryActionJob($actionJob: ID!) {
                retryActionJob(actionJob: $actionJob) {
                    alwaysNil
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
                        executionStart
                        executionEnd
                        patchSet {
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
                                invokationReason
                                status {
                                    errors
                                    state
                                    pendingCount
                                    completedCount
                                }
                                patchSet {
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
        map((data) => data.actions)
    )

export const fetchActionByID = (action: ID): Observable<IAction | null> =>
    queryGraphQL(
        gql`
            query ActionByID($action: ID!) {
                node(id: $action) {
                    __typename
                    ... on Action {
                        id
                        savedSearch {
                            description
                        }
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
                        schedule
                        actionExecutions {
                            totalCount
                            nodes {
                                id
                                invokationReason
                                status {
                                    errors
                                    state
                                    pendingCount
                                    completedCount
                                }
                                patchSet {
                                    id
                                }
                            }
                        }
                    }
                }
            }
        `,
        { action }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                return null
            }
            if (node.__typename !== 'Action') {
                throw new Error(`The given ID is a ${node.__typename}, not a Action`)
            }
            return node
        })
    )

export async function createActionExecution(actionID: ID): Promise<IActionExecution> {
    const result = await mutateGraphQL(
        gql`
            mutation CreateActionExecution($action: ID!) {
                createActionExecution(action: $action) {
                    id
                }
            }
        `,
        { action: actionID }
    ).toPromise()
    return dataOrThrowErrors(result).createActionExecution
}

export async function createAction(definition: string): Promise<IAction> {
    const result = await mutateGraphQL(
        gql`
            mutation CreateAction($definition: JSONCString!) {
                createAction(definition: $definition) {
                    id
                }
            }
        `,
        { definition }
    ).toPromise()
    return dataOrThrowErrors(result).createAction
}

export async function updateAction(action: ID, newDefinition: string): Promise<IAction> {
    const result = await mutateGraphQL(
        gql`
            mutation UpdateAction($action: ID!, $newDefinition: JSONCString!) {
                updateAction(action: $action, newDefinition: $newDefinition) {
                    id
                }
            }
        `,
        { newDefinition, action }
    ).toPromise()
    return dataOrThrowErrors(result).updateAction
}

export async function updateActionJob(actionJob: ID, { state }: { state?: ActionJobState }): Promise<IActionJob> {
    const result = await mutateGraphQL(
        gql`
            mutation UpdateActionJob($actionJob: ID!, $state: ActionJobState) {
                updateActionJob(actionJob: $actionJob, state: $state) {
                    id
                }
            }
        `,
        { state, actionJob }
    ).toPromise()
    return dataOrThrowErrors(result).updateActionJob
}
