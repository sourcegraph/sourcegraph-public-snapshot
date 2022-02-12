import React from 'react'

import { dataOrThrowErrors } from '@sourcegraph/http-client'
import { Card } from '@sourcegraph/wildcard'

import { useConnection } from '../../components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
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
    const pageSize = 20
    const runPageSize = 20

    const { connection, error, loading, fetchMore, hasNextPage } = useConnection<
        MonitorTriggerEventsResult,
        MonitorTriggerEventsVariables,
        CodeMonitorWithEvents
    >({
        query: ListCodeMonitorsWithEventsQuery,
        variables: { first: pageSize, after: null, triggerEventsFirst: runPageSize, triggerEventsAfter: null },
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
            <p>
                {/* TODO: Text to change */}
                You can use these logs to troubleshoot issues with code monitor notifications. Only the {
                    runPageSize
                }{' '}
                most recent runs are shown and old runs are deleted periodically.
            </p>
            <Card className="px-3 pt-3">
                <ConnectionContainer>
                    {error && <ConnectionError errors={[error.message]} />}
                    <ConnectionList className={styles.grid}>
                        {connection?.nodes.length ? <CodeMonitorLogsHeader /> : null}
                        {connection?.nodes.map(monitor => (
                            <MonitorLogNode key={monitor.id} monitor={monitor} />
                        ))}
                    </ConnectionList>
                    {loading && <ConnectionLoading />}
                    {connection && (
                        <SummaryContainer centered={true}>
                            <ConnectionSummary
                                noSummaryIfAllNodesVisible={true}
                                first={pageSize}
                                connection={connection}
                                noun="monitor"
                                pluralNoun="monitors"
                                hasNextPage={hasNextPage}
                                emptyElement={<div>You haven't created any monitors yet</div>}
                            />
                            {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                        </SummaryContainer>
                    )}
                </ConnectionContainer>
            </Card>
        </div>
    )
}
