import { type FC, useEffect } from 'react'

import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import { Container } from '@sourcegraph/wildcard'

import {
    PermissionsSyncJobsSearchType,
    type PermissionsSyncJobsStatsResult,
    type PermissionsSyncJobsStatsVariables,
    PermissionsSyncJobState,
} from '../../graphql-operations'
import { ValueLegendItem } from '../analytics/components/ValueLegendList'

import { PERMISSIONS_SYNC_JOBS_STATS } from './backend'
import { type Filters, PERMISSIONS_SYNC_JOBS_POLL_INTERVAL } from './PermissionsSyncJobsTable'

import styles from './styles.module.scss'

interface Props {
    polling: boolean
    filters: Filters
    setFilters: (data: Partial<Filters>) => void
}

export const PermissionsSyncStats: FC<Props> = ({ polling, filters, setFilters }) => {
    const { data, startPolling, stopPolling } = useQuery<
        PermissionsSyncJobsStatsResult,
        PermissionsSyncJobsStatsVariables
    >(PERMISSIONS_SYNC_JOBS_STATS, {})

    useEffect(() => {
        if (polling) {
            startPolling(PERMISSIONS_SYNC_JOBS_POLL_INTERVAL)
        } else {
            stopPolling()
        }
    }, [polling, stopPolling, startPolling])

    return (
        <Container className="mb-1">
            {data && (
                <div className={styles.statsBox}>
                    <div className="d-flex">
                        <ValueLegendItem
                            value={data.permissionsSyncingStats?.queueSize}
                            className={classNames(styles.stat)}
                            description="Queued"
                            color="var(--body-color)"
                            tooltip="The number of permissions sync jobs in the queue."
                        />
                        <ValueLegendItem
                            value={data.permissionsSyncingStats?.usersWithLatestJobFailing}
                            secondValue={data.site?.users.totalCount}
                            className={classNames(styles.stat)}
                            description="Failing users"
                            color="var(--body-color)"
                            tooltip="The number of users with latest permissions sync job failing."
                            onClick={() =>
                                setFilters({
                                    ...filters,
                                    state: PermissionsSyncJobState.FAILED,
                                    searchType: PermissionsSyncJobsSearchType.USER,
                                })
                            }
                        />
                        <ValueLegendItem
                            value={data.permissionsSyncingStats?.usersWithNoPermissions}
                            secondValue={data.site?.users.totalCount}
                            className={classNames(styles.stat)}
                            description="No perms users"
                            color="var(--body-color)"
                            tooltip="The number of users with no permissions."
                        />
                        <ValueLegendItem
                            value={data.permissionsSyncingStats?.usersWithStalePermissions}
                            secondValue={data.site?.users.totalCount}
                            className={classNames(styles.stat)}
                            description="Outdated users"
                            color="var(--body-color)"
                            tooltip="The number of users with old permissions."
                        />
                        <ValueLegendItem
                            value={data.permissionsSyncingStats?.reposWithLatestJobFailing}
                            secondValue={data.repositoryStats?.total}
                            className={classNames(styles.stat)}
                            description="Failing repos"
                            color="var(--body-color)"
                            tooltip="The number of repos with latest permissions sync job failing."
                            onClick={() =>
                                setFilters({
                                    ...filters,
                                    state: PermissionsSyncJobState.FAILED,
                                    searchType: PermissionsSyncJobsSearchType.REPOSITORY,
                                })
                            }
                        />
                        <ValueLegendItem
                            value={data.permissionsSyncingStats?.reposWithNoPermissions}
                            secondValue={data.repositoryStats?.total}
                            className={classNames(styles.stat)}
                            description="No perms repos"
                            color="var(--body-color)"
                            tooltip="The number of repos with no permissions."
                        />
                        <ValueLegendItem
                            value={data.permissionsSyncingStats?.reposWithStalePermissions}
                            secondValue={data.repositoryStats?.total}
                            className={classNames(styles.stat)}
                            description="Outdated repos"
                            color="var(--body-color)"
                            tooltip="The number of repos with old permissions."
                        />
                    </div>
                </div>
            )}
        </Container>
    )
}
