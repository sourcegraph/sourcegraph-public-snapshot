import React from 'react'

import { dataOrThrowErrors } from '@sourcegraph/http-client'
import { Card } from '@sourcegraph/wildcard'

import { useConnection } from '../../components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
} from '../../components/FilteredConnection/ui'
import {
    CodeMonitorWithEvents,
    MonitorTriggerEventsResult,
    MonitorTriggerEventsVariables,
} from '../../graphql-operations'

import { ListCodeMonitorsWithEventsQuery } from './backend'
import styles from './CodeMonitoringLogs.module.scss'
import { CodeMonitorLogsHeader } from './components/logs/CodeMonitorLogsHeader'
import { MonitorLogNode } from './components/logs/MonitorLogNode'

export const CodeMonitoringLogs: React.FunctionComponent<{}> = () => {
    const { connection, error, loading, fetchMore, hasNextPage } = useConnection<
        MonitorTriggerEventsResult,
        MonitorTriggerEventsVariables,
        CodeMonitorWithEvents
    >({
        query: ListCodeMonitorsWithEventsQuery,
        variables: { first: 20, after: null, triggerEventsFirst: 20, triggerEventsAfter: null },
        getConnection: result => {
            const data = dataOrThrowErrors(result)

            if (!data.currentUser) {
                throw new Error('User is not logged in')
            }
            return data.currentUser.monitors
        },
    })

    return (
        <div>
            <h2>Code Monitoring Logs</h2>
            <Card className="p-3">
                <ConnectionContainer>
                    {error && <ConnectionError errors={[error.message]} />}
                    <ConnectionList className={styles.grid}>
                        {connection?.nodes.length ? <CodeMonitorLogsHeader /> : null}
                        {connection?.nodes.map(monitor => (
                            <MonitorLogNode key={monitor.id} monitor={monitor} />
                        ))}
                    </ConnectionList>
                    {loading && <ConnectionLoading />}
                </ConnectionContainer>
            </Card>
        </div>
    )
}
