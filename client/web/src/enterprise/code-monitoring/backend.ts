import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import { requestGraphQL } from '../../backend/graphql'
import {
    CreateCodeMonitorResult,
    CreateCodeMonitorVariables,
    ListCodeMonitors,
    ListUserCodeMonitorsResult,
    ListUserCodeMonitorsVariables,
    ToggleCodeMonitorEnabledResult,
    ToggleCodeMonitorEnabledVariables,
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
        actions {
            nodes {
                ... on MonitorEmail {
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

export const listUserCodeMonitors = ({
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
