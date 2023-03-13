import { FC, useMemo } from 'react'

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
import { REPOSITORY_STATS } from '../../site-admin/backend'

import styles from './ProgressBar.module.scss'

export const ProgressBar: FC<{}> = () => {
    const { data } = useQuery<RepositoryStatsResult, RepositoryStatsVariables>(REPOSITORY_STATS, { pollInterval: 2000 })

    const { data: statusData } = useQuery<StatusMessagesResult>(STATUS_MESSAGES, {
        fetchPolicy: 'no-cache',
        pollInterval: 5000,
    })

    const formatNumber = (num: string | number): string => num.toLocaleString('en-US')

    const items = useMemo(
        () => [
            {
                value: data?.repositoryStats.total,
                description: 'Repositories',
                color: 'text-merged',
            },
            {
                value: data?.repositoryStats.notCloned,
                description: 'Not cloned',
            },
            {
                value: data?.repositoryStats.cloning,
                description: 'Cloning',
            },
            {
                value: data?.repositoryStats.cloned,
                description: 'Cloned',
                color: 'text-success',
            },
            {
                value: data?.repositoryStats.indexed,
                description: 'Indexed',
            },
            {
                value: data?.repositoryStats.failedFetch,
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
                    <span className={classNames('font-weight-bold', item?.color)}>{formatNumber(item.value ?? 0)}</span>{' '}
                    {item.description}
                </Text>
            ))}
        </section>
    )
}
