import { FC, useEffect, useMemo } from 'react'

import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import {
    CloudAlertIconRefresh,
    CloudCheckIconRefresh,
    CloudSyncIconRefresh,
} from '@sourcegraph/shared/src/components/icons'
import { Icon, Text } from '@sourcegraph/wildcard'

import { RepositoryStatsResult, RepositoryStatsVariables, StatusMessagesResult } from '../../graphql-operations'

import { REPOSITORY_STATS, REPO_PAGE_POLL_INTERVAL } from '../../site-admin/backend'
import { STATUS_MESSAGES } from '../../nav/StatusMessagesNavItemQueries'

// TODO: Dynamic header title

import styles from './ProgressBar.module.scss'

export const ProgressBar: FC<{}> = () => {
    const { data, startPolling, stopPolling } = useQuery<RepositoryStatsResult, RepositoryStatsVariables>(
        REPOSITORY_STATS,
        {}
    )

    const { data: statusData } = useQuery<StatusMessagesResult>(STATUS_MESSAGES, {
        fetchPolicy: 'no-cache',
        pollInterval: 10000,
    })

    useEffect(() => {
        if (data?.repositoryStats?.total === 0 || data?.repositoryStats?.cloning !== 0) {
            startPolling(REPO_PAGE_POLL_INTERVAL)
        } else {
            stopPolling()
        }
    }, [data, startPolling, stopPolling])

    const formatNumber = (num: string | number): string => {
        return num.toLocaleString('en-US')
    }

    const statusMessage: JSX.Element = useMemo(() => {
        let codeHostMessage
        let iconProps

        if (
            !statusData ||
            statusData.statusMessages?.some(
                ({ __typename: type }) => type === 'CloningProgress' || type === 'IndexingProgress'
            )
        ) {
            codeHostMessage = 'Syncing'
            iconProps = { as: CloudSyncIconRefresh }
        } else if (
            statusData.statusMessages?.some(
                ({ __typename: type }) =>
                    type === 'GitUpdatesDisabled' || type === 'ExternalServiceSyncError' || type === 'SyncError'
            )
        ) {
            // TODO: Error handle. What action do we allow if there's an error & no synced repos?
            codeHostMessage = 'Error'
            iconProps = { as: CloudAlertIconRefresh }
        } else {
            codeHostMessage = 'Synced'
            iconProps = { as: CloudCheckIconRefresh }
        }

        return (
            <div className="d-flex align-items-center mr-2">
                <Icon {...iconProps} size="md" aria-label={codeHostMessage} className="mr-1" />
                <Text className={classNames(codeHostMessage === 'Syncing' && styles.loading, 'mb-0')} size="small">
                    {codeHostMessage}
                </Text>
            </div>
        )
    }, [statusData])

    if (data?.repositoryStats.total === 0) {
        return null
    }

    return (
        <section className="d-flex align-items-center py-1">
            {statusMessage}

            <Text className="mb-0 mr-3" size="small">
                <span className="font-weight-bold text-merged">{formatNumber(data?.repositoryStats.total ?? 0)}</span>{' '}
                Repositories
            </Text>
            <Text className="mb-0 mr-3" size="small">
                <span className="font-weight-bold">{formatNumber(data?.repositoryStats.notCloned ?? 0)}</span> Not
                cloned
            </Text>
            <Text className="mb-0 mr-3" size="small">
                <span className="font-weight-bold">{formatNumber(data?.repositoryStats.cloning ?? 0)}</span> Cloning
            </Text>
            <Text className="mb-0 mr-3" size="small">
                <span className="font-weight-bold text-success">{formatNumber(data?.repositoryStats.cloned ?? 0)}</span>{' '}
                Cloned
            </Text>
            <Text className="mb-0 mr-3" size="small">
                <span className="font-weight-bold">{formatNumber(data?.repositoryStats.indexed ?? 0)}</span> Indexed
            </Text>
            <Text className="mb-0" size="small">
                <span className="font-weight-bold text-danger">
                    {formatNumber(data?.repositoryStats.failedFetch ?? 0)}
                </span>{' '}
                Failed
            </Text>
        </section>
    )
}
