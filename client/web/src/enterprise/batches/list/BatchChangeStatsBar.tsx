import React from 'react'

import { mdiInformationOutline } from '@mdi/js'
import classNames from 'classnames'

import { pluralize } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { H3, H4, Icon, LoadingSpinner, Tooltip } from '@sourcegraph/wildcard'

import type { GlobalChangesetsStatsResult, GlobalChangesetsStatsVariables } from '../../../graphql-operations'
import { DEFAULT_MINS_SAVED_PER_CHANGESET } from '../../../site-admin/analytics/AnalyticsBatchChangesPage'
import { ChangesetStatusClosed, ChangesetStatusOpen } from '../detail/changesets/ChangesetStatusCell'

import { GLOBAL_CHANGESETS_STATS } from './backend'

import styles from './BatchChangeStatsBar.module.scss'

interface BatchChangeStatsBarProps {
    className?: string
}

export const BatchChangeStatsBar: React.FunctionComponent<React.PropsWithChildren<BatchChangeStatsBarProps>> = () => {
    const [minSavedPerChangeset = DEFAULT_MINS_SAVED_PER_CHANGESET] =
        useTemporarySetting('batches.minSavedPerChangeset')

    const { data, loading } = useQuery<GlobalChangesetsStatsResult, GlobalChangesetsStatsVariables>(
        GLOBAL_CHANGESETS_STATS,
        {}
    )

    if (loading && !data) {
        return (
            <div className="text-center">
                <LoadingSpinner className="mx-auto my-4" />
            </div>
        )
    }
    if (!data) {
        return null
    }

    return (
        <div className={classNames(styles.statsBar, 'text-muted d-block d-sm-flex')}>
            <div className={styles.leftSide}>
                <div className="pr-4">
                    <H3 className="font-weight-bold mb-0">
                        <span className="d-block mb-1">{data.batchChanges.totalCount}</span>
                        <span className={styles.statLabel}>Batch changes</span>
                    </H3>
                </div>
                <div className="pr-4">
                    <H3 className="font-weight-bold mb-0">
                        <span className="d-block mb-1">{data.globalChangesetsStats.merged}</span>
                        <span className={styles.statLabel}>Changesets merged</span>
                    </H3>
                </div>
                <div className="pr-4">
                    <H3 className="font-weight-bold mb-0">
                        <span className="d-block mb-1">
                            {Math.round((data.globalChangesetsStats.merged * minSavedPerChangeset) / 60).toFixed(2)}
                        </span>
                        <span className={styles.statLabel}>Hours saved</span>
                        <Tooltip content="Based on multiplier per changeset defined by site admin">
                            <Icon
                                aria-label="Based on multiplier per changeset defined by site admin"
                                svgPath={mdiInformationOutline}
                                className="ml-1"
                            />
                        </Tooltip>
                    </H3>
                </div>
            </div>
            <div className={styles.rightSide}>
                <div className="pr-4 text-center">
                    <ChangesetStatusOpen
                        className="d-flex"
                        label={
                            <H4
                                className="font-weight-normal m-0"
                                aria-label={`${data.globalChangesetsStats.open} total ${pluralize(
                                    'changeset',
                                    data.globalChangesetsStats.open
                                )} open`}
                            >
                                {data.globalChangesetsStats.open} open
                            </H4>
                        }
                    />
                </div>
                <div className="text-center">
                    <ChangesetStatusClosed
                        className="d-flex"
                        label={
                            <H4
                                className="font-weight-normal m-0"
                                aria-label={`${data.globalChangesetsStats.closed} total ${pluralize(
                                    'changeset',
                                    data.globalChangesetsStats.closed
                                )} closed`}
                            >
                                {data.globalChangesetsStats.closed} closed
                            </H4>
                        }
                    />
                </div>
            </div>
        </div>
    )
}
