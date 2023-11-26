import React, { useMemo } from 'react'

import classNames from 'classnames'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { Card, Text } from '@sourcegraph/wildcard'

import { useShowMorePagination } from '../../components/FilteredConnection/hooks/useShowMorePagination'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../components/FilteredConnection/ui'
import type {
    CodeMonitorWithEvents,
    MonitorTriggerEventsResult,
    MonitorTriggerEventsVariables,
} from '../../graphql-operations'

import { CodeMonitorLogsHeader } from './components/logs/CodeMonitorLogsHeader'
import { MonitorLogNode } from './components/logs/MonitorLogNode'

import styles from './CodeMonitoringLogs.module.scss'

export const CODE_MONITOR_EVENTS = gql`
    query MonitorTriggerEvents($first: Int, $after: String, $triggerEventsFirst: Int, $triggerEventsAfter: String) {
        currentUser {
            monitors(first: $first, after: $after) {
                nodes {
                    ...CodeMonitorWithEvents
                }
                totalCount
                pageInfo {
                    endCursor
                    hasNextPage
                }
            }
        }
    }

    fragment CodeMonitorTriggerEvents on MonitorQuery {
        events(first: $triggerEventsFirst, after: $triggerEventsAfter) {
            nodes {
                ...MonitorTriggerEventWithActions
            }
            totalCount
            pageInfo {
                endCursor
                hasNextPage
            }
        }
    }

    fragment MonitorTriggerEventWithActions on MonitorTriggerEvent {
        id
        status
        message
        timestamp
        resultCount
        query
        actions {
            nodes {
                ... on MonitorWebhook {
                    __typename
                    events {
                        ...MonitorActionEvents
                    }
                }
                ... on MonitorEmail {
                    __typename
                    events {
                        ...MonitorActionEvents
                    }
                }
                ... on MonitorSlackWebhook {
                    __typename
                    events {
                        ...MonitorActionEvents
                    }
                }
            }
        }
    }

    fragment CodeMonitorWithEvents on Monitor {
        description
        id
        trigger {
            ... on MonitorQuery {
                query
                ...CodeMonitorTriggerEvents
            }
        }
    }

    fragment MonitorActionEvents on MonitorActionEventConnection {
        nodes {
            id
            status
            message
            timestamp
        }
    }
`

export const CodeMonitoringLogs: React.FunctionComponent<
    React.PropsWithChildren<{ now?: () => Date; _testStartOpen?: boolean }>
> = ({
    now,
    _testStartOpen = false, // For testing purposes only; force everything to start expanded
}) => {
    const pageSize = 20
    const runPageSize = 20

    const { connection, error, loading, fetchMore, hasNextPage } = useShowMorePagination<
        MonitorTriggerEventsResult,
        MonitorTriggerEventsVariables,
        CodeMonitorWithEvents
    >({
        query: CODE_MONITOR_EVENTS,
        variables: { first: pageSize, after: null, triggerEventsFirst: runPageSize, triggerEventsAfter: null },
        getConnection: result => {
            const data = dataOrThrowErrors(result)

            if (!data.currentUser) {
                throw new Error('User is not logged in')
            }
            return data.currentUser.monitors
        },
    })

    const monitors: CodeMonitorWithEvents[] = useMemo(() => connection?.nodes ?? [], [connection])

    return (
        <div>
            <Text>
                Use these logs to troubleshoot issues with code monitor notifications. Only the {runPageSize} most
                recent runs are shown. Old runs are deleted periodically.
            </Text>
            <Card className="p-3">
                <ConnectionContainer>
                    {error && <ConnectionError errors={[error.message]} />}
                    {monitors.length > 0 ? <CodeMonitorLogsHeader /> : null}
                    <ConnectionList className={classNames('mb-0')}>
                        {monitors.map(monitor => (
                            <MonitorLogNode key={monitor.id} monitor={monitor} now={now} startOpen={_testStartOpen} />
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
                                emptyElement={
                                    <div className={styles.empty}>No code monitors have been created yet.</div>
                                }
                            />
                            {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                        </SummaryContainer>
                    )}
                </ConnectionContainer>
            </Card>
        </div>
    )
}
