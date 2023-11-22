import { type FC, useMemo } from 'react'

import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import {
    CloudAlertIconRefresh,
    CloudCheckIconRefresh,
    CloudSyncIconRefresh,
} from '@sourcegraph/shared/src/components/icons'
import { Icon, Text } from '@sourcegraph/wildcard'

import type { StatusAndRepoStatsResult } from '../../graphql-operations'
import { STATUS_AND_REPO_STATS } from '../../site-admin/backend'

import styles from './ProgressBar.module.scss'

export const ProgressBar: FC<{}> = () => {
    const { data } = useQuery<StatusAndRepoStatsResult>(STATUS_AND_REPO_STATS, {
        fetchPolicy: 'cache-and-network',
        pollInterval: 2000,
    })

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
                description: 'Queued',
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

        if (!data || data.statusMessages?.some(({ __typename: type }) => type === 'CloningProgress')) {
            codeHostMessage = 'Syncing'
            iconProps = { as: CloudSyncIconRefresh }
        } else if (data.statusMessages?.some(({ __typename: type }) => type === 'IndexingProgress')) {
            codeHostMessage = 'Indexing'
            iconProps = { as: CloudSyncIconRefresh }
        } else if (
            data.statusMessages?.some(
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
                <Text
                    className={classNames(
                        (codeHostMessage === 'Syncing' || codeHostMessage === 'Indexing') && styles.loading,
                        'mb-0'
                    )}
                    size="small"
                >
                    {codeHostMessage}
                </Text>
            </div>
        )
    }, [data])

    const totalRepositories = data?.repositoryStats.total ?? 0

    // If there is no repositories do not render progress bar UI
    if (totalRepositories === 0) {
        return null
    }

    return (
        <section className={styles.root}>
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
