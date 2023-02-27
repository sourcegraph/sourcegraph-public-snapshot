import { FC, useEffect } from 'react'

import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import { CloudSyncIconRefresh } from '@sourcegraph/shared/src/components/icons'
import { Icon, Text } from '@sourcegraph/wildcard'

import { RepositoryStatsResult, RepositoryStatsVariables } from '../../graphql-operations'

import { REPOSITORY_STATS, REPO_PAGE_POLL_INTERVAL } from '../../site-admin/backend'

interface ProgressBarProps {}

// TODO: loading/error?, d-none if no stats
// TODO: Dynamic header title

export const ProgressBar: FC<ProgressBarProps> = props => {
    const { data, loading, error, startPolling, stopPolling } = useQuery<
        RepositoryStatsResult,
        RepositoryStatsVariables
    >(REPOSITORY_STATS, {})

    useEffect(() => {
        console.log(data)
        if (data?.repositoryStats?.total === 0 || data?.repositoryStats?.cloning !== 0) {
            startPolling(REPO_PAGE_POLL_INTERVAL)
        } else {
            stopPolling()
        }
    }, [data, startPolling, stopPolling])

    // TODO: Format #s, grid space items
    return (
        <section className={classNames(data?.repositoryStats.total === 0 ? 'd-none' : 'd-flex py-1')}>
            <div className="d-flex align-items-center mr-3">
                <Icon as={CloudSyncIconRefresh} size="sm" className="mr-1" />
                <Text className="mb-0" size="small">
                    Syncing...
                </Text>
            </div>
            <Text className="mb-0 mr-3" size="small">
                <span className="font-weight-bold text-merged">
                    {data?.repositoryStats.total.toLocaleString('en-US')}
                </span>{' '}
                Repositories
            </Text>
            <Text className="mb-0 mr-3" size="small">
                <span className="font-weight-bold">{data?.repositoryStats.notCloned}</span> Not cloned
            </Text>
            <Text className="mb-0 mr-3" size="small">
                <span className="font-weight-bold">{data?.repositoryStats.cloning}</span> Cloning
            </Text>
            <Text className="mb-0 mr-3" size="small">
                <span className="font-weight-bold text-success">{data?.repositoryStats.cloned}</span> Cloned
            </Text>
            <Text className="mb-0 mr-3" size="small">
                <span className="font-weight-bold">{data?.repositoryStats.indexed}</span> Indexed
            </Text>
            <Text className="mb-0 mr-3" size="small">
                <span className="font-weight-bold text-error">{data?.repositoryStats.failedFetch}</span> Failed
            </Text>
        </section>
    )
}
