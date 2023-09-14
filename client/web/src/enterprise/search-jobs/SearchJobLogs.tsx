import { FC, useLayoutEffect, useState } from 'react'

import classNames from 'classnames'
import { timeFormat } from 'd3-time-format'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { ErrorAlert, Input, LoadingSpinner, Text, Button } from '@sourcegraph/wildcard'

import { useShowMorePagination } from '../../components/FilteredConnection/hooks/useShowMorePagination'
import { ListPageZeroState } from '../../components/ZeroStates/ListPageZeroState'
import { GetSearchJobLogsResult, GetSearchJobLogsVariables } from '../../graphql-operations'

import styles from './SearchJobLogs.module.scss'

const formatDate = timeFormat('%H:%M:%S')

/**
 * Main query to fetch list of search job logs, exported only for Storybook story
 * apollo mocks, not designed to be reused in other places.
 */
export const SEARCH_JOB_LOGS = gql`
    query GetSearchJobLogs($id: ID!, $first: Int!, $after: String) {
        searchJob(id: $id) {
            logs(first: $first, after: $after) {
                nodes {
                    text
                    time
                }
                pageInfo {
                    hasNextPage
                    endCursor
                }
                totalCount
            }
        }
    }
`

// Derived search job log type since gql type generator doesn't
// include all standard gql types from schema (only something that is used
// in the query itself)
type SearchJobLog = GetSearchJobLogsResult['searchJob']['logs']['nodes'][number]

interface SearchJobLogsProps {
    jobId: string
    onContentChange: () => void
}

export const SearchJobLogs: FC<SearchJobLogsProps> = props => {
    const { jobId, onContentChange } = props

    const [searchTerm, setSearchTerm] = useState('')

    const { connection, error, hasNextPage, loading, fetchMore } = useShowMorePagination<
        GetSearchJobLogsResult,
        GetSearchJobLogsVariables,
        SearchJobLog
    >({
        query: SEARCH_JOB_LOGS,
        variables: {
            id: jobId,
            first: 50,
            after: null,
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)

            return data?.searchJob.logs
        },
    })

    // Observe any changes about async query and trigger onContentChange
    // handler in order to update popover position based on new content
    // from the query.
    useLayoutEffect(() => {
        onContentChange()
    }, [onContentChange, connection, loading, error])

    return (
        <div className={styles.root}>
            <header className={styles.header}>
                <Input
                    value={searchTerm}
                    disabled={!!error || connection?.nodes.length === 0}
                    placeholder="Filter search job logs..."
                    onChange={event => setSearchTerm(event.target.value)}
                />
            </header>

            {error && !loading && <ErrorAlert error={error} className={styles.error} />}

            {loading && !connection && (
                <Text className={styles.loading}>
                    <LoadingSpinner /> Fetching search job logs
                </Text>
            )}

            {!error && connection && (
                <ul className={styles.logs}>
                    {connection.nodes.length === 0 && <SearchJobLogsZeroState withFilters={searchTerm.length > 0} />}
                    {connection.nodes.map(log => (
                        <li key={log.time} className={styles.log}>
                            <Text className={styles.logTime}>[{formatDate(new Date(log.time))}]</Text>
                            <Text className={classNames(styles.logText, 'text-monospace')}>{log.text}</Text>
                        </li>
                    ))}

                    {hasNextPage && (
                        <li className={styles.loadingMore}>
                            <Button variant="link" size="sm" onClick={() => fetchMore()}>
                                Show more logs
                            </Button>
                        </li>
                    )}
                </ul>
            )}
        </div>
    )
}

interface SearchJobLogsZeroStateProps {
    withFilters: boolean
}

const SearchJobLogsZeroState: FC<SearchJobLogsZeroStateProps> = props => (
    <ListPageZeroState
        size="small"
        title={props.withFilters ? 'No logs found' : 'No logs found yet'}
        subTitle={
            props.withFilters ? (
                'Try to reset filter to see all search jobs logs'
            ) : (
                <>
                    Search job hasn't produced any logs yet,
                    <br /> try to check it later
                </>
            )
        }
        className={styles.zeroState}
    />
)
