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
import { STATUS_MESSAGES } from '../../nav/StatusMessagesNavItemQueries'
import { REPOSITORY_STATS, REPO_PAGE_POLL_INTERVAL } from '../../site-admin/backend'

import styles from './ProgressBar.module.scss'

export const ProgressBar: FC<{}> = () => {
    const { data, startPolling, stopPolling } = useQuery<RepositoryStatsResult, RepositoryStatsVariables>(
        REPOSITORY_STATS,
        {}
    )

    const { data: statusData } = useQuery<StatusMessagesResult>(STATUS_MESSAGES, {
        fetchPolicy: 'no-cache',
        pollInterval: 5000,
    })

    useEffect(() => {
        if (data?.repositoryStats?.total === 0 || data?.repositoryStats?.cloning !== 0) {
            startPolling(REPO_PAGE_POLL_INTERVAL)
        } else {
            stopPolling()
        }
    }, [data, startPolling, stopPolling])

    const formatNumber = (num: string | number): string => num.toLocaleString('en-US')

    const items = useMemo(
        () => [
            {
                value: Math.max(data?.repositoryStats.total ?? 0, 0),
                description: 'Repositories',
                color: 'text-merged',
            },
            {
                value: Math.max(data?.repositoryStats.notCloned ?? 0, 0),
                description: 'Not cloned',
            },
            {
                value: Math.max(data?.repositoryStats.cloning ?? 0, 0),
                description: 'Cloning',
            },
            {
                value: Math.max(data?.repositoryStats.cloned ?? 0, 0),
                description: 'Cloned',
                color: 'text-success',
            },
            {
                value: Math.max(data?.repositoryStats.indexed ?? 0, 0),
                description: 'Indexed',
            },
            {
                value: Math.max(data?.repositoryStats.failedFetch ?? 0, 0),
                description: 'Failed',
                color: 'text-danger',
            },
        ],
        [data]
    )

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
        <section className="d-flex align-items-center">
            {statusMessage}

            {items.map(item => (
                <Text className="mb-0 mr-3" size="small" key={item.description}>
                    <span className={classNames('font-weight-bold', item?.color)}>{formatNumber(item.value)}</span>{' '}
                    {item.description}
                </Text>
            ))}
        </section>
    )
}
