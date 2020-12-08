import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import { requestGraphQL } from '../../backend/graphql'
import {
    CreateCodeMonitorResult,
    CreateCodeMonitorVariables,
    FetchCodeMonitorResult,
    FetchCodeMonitorVariables,
    ListCodeMonitors,
    ListUserCodeMonitorsResult,
    ListUserCodeMonitorsVariables,
    MonitorEditActionInput,
    MonitorEditInput,
    MonitorEditTriggerInput,
    ToggleCodeMonitorEnabledResult,
    ToggleCodeMonitorEnabledVariables,
    UpdateCodeMonitorResult,
    UpdateCodeMonitorVariables,
} from '../../graphql-operations'

export const createCodeMonitor = ({
    monitor,
    trigger,
    actions,
}: CreateCodeMonitorVariables): Observable<CreateCodeMonitorResult['createCodeMonitor']> => {
    const query = gql`
        mutation CreateCodeMonitor(
            $monitor: MonitorInput!
            $trigger: MonitorTriggerInput!
            $actions: [MonitorActionInput!]!
        ) {
            createCodeMonitor(monitor: $monitor, trigger: $trigger, actions: $actions) {
                description
            }
        }
    `

    return requestGraphQL<CreateCodeMonitorResult, CreateCodeMonitorVariables>(query, {
        monitor,
        trigger,
        actions,
    }).pipe(
        map(dataOrThrowErrors),
        map(data => data.createCodeMonitor)
    )
}

const CodeMonitorFragment = gql`
    fragment CodeMonitorFields on Monitor {
        id
        description
        enabled
        trigger {
            ... on MonitorQuery {
                id
                query
            }
        }
        actions {
            nodes {
                ... on MonitorEmail {
                    id
                    enabled
                    recipients {
                        nodes {
                            id
                        }
                    }
                }
            }
        }
    }
`

const ListCodeMonitorsFragment = gql`
    fragment ListCodeMonitors on MonitorConnection {
        nodes {
            ...CodeMonitorFields
        }
        totalCount
        pageInfo {
            endCursor
            hasNextPage
        }
    }
    ${CodeMonitorFragment}
`

export interface ListCodeMonitorsResult {
    monitors: ListCodeMonitors
}

export const fetchUserCodeMonitors = ({
    id,
    first,
    after,
}: ListUserCodeMonitorsVariables): Observable<ListCodeMonitors> => {
    const query = gql`
        query ListUserCodeMonitors($id: ID!, $first: Int, $after: String) {
            node(id: $id) {
                __typename
                ... on User {
                    monitors(first: $first, after: $after) {
                        ...ListCodeMonitors
                    }
                }
            }
        }
        ${ListCodeMonitorsFragment}
    `

    return requestGraphQL<ListUserCodeMonitorsResult, ListUserCodeMonitorsVariables>(query, {
        id,
        first,
        after,
    }).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.node) {
                throw new Error('namespace not found')
            }

            if (data.node.__typename !== 'User') {
                throw new Error(`Requested node is a ${data.node.__typename}, not a User or Org`)
            }

            return data.node.monitors
        })
    )
}

export const toggleCodeMonitorEnabled = (
    id: string,
    enabled: boolean
): Observable<ToggleCodeMonitorEnabledResult['toggleCodeMonitor']> => {
    const query = gql`
        mutation ToggleCodeMonitorEnabled($id: ID!, $enabled: Boolean!) {
            toggleCodeMonitor(id: $id, enabled: $enabled) {
                id
                enabled
            }
        }
    `

    return requestGraphQL<ToggleCodeMonitorEnabledResult, ToggleCodeMonitorEnabledVariables>(query, {
        id,
        enabled,
    }).pipe(
        map(dataOrThrowErrors),
        map(data => data.toggleCodeMonitor)
    )
}

export const fetchCodeMonitor = (id: string): Observable<FetchCodeMonitorResult> => {
    const query = gql`
        query FetchCodeMonitor($id: ID!) {
            node(id: $id) {
                ... on Monitor {
                    id
                    description
                    owner {
                        id
                        namespaceName
                    }
                    enabled
                    actions {
                        nodes {
                            ... on MonitorEmail {
                                id
                                recipients {
                                    nodes {
                                        id
                                        url
                                    }
                                }
                                enabled
                            }
                        }
                    }
                    trigger {
                        ... on MonitorQuery {
                            id
                            query
                        }
                    }
                }
            }
        }
    `

    return requestGraphQL<FetchCodeMonitorResult, FetchCodeMonitorVariables>(query, {
        id,
    }).pipe(
        map(dataOrThrowErrors),
        map(data => data)
    )
}

export const updateCodeMonitor = (
    monitorEditInput: MonitorEditInput,
    triggerEditInput: MonitorEditTriggerInput,
    actionEditInput: MonitorEditActionInput[]
): Observable<UpdateCodeMonitorResult['updateCodeMonitor']> => {
    const updateCodeMonitorQuery = gql`
        mutation UpdateCodeMonitor(
            $monitor: MonitorEditInput!
            $trigger: MonitorEditTriggerInput!
            $actions: [MonitorEditActionInput!]!
        ) {
            updateCodeMonitor(monitor: $monitor, trigger: $trigger, actions: $actions) {
                ...CodeMonitorFields
            }
        }
        ${CodeMonitorFragment}
    `

    return requestGraphQL<UpdateCodeMonitorResult, UpdateCodeMonitorVariables>(updateCodeMonitorQuery, {
        monitor: monitorEditInput,
        trigger: triggerEditInput,
        actions: actionEditInput,
    }).pipe(
        map(dataOrThrowErrors),
        map(data => data.updateCodeMonitor)
    )
}
