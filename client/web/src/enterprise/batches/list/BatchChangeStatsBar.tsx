import React from 'react'

import { mdiInformationOutline } from '@mdi/js'

import { useQuery } from '@sourcegraph/http-client'
import { Icon, LoadingSpinner, Tooltip } from '@sourcegraph/wildcard'

import { GlobalChangesetsStatsResult, GlobalChangesetsStatsVariables } from '../../../graphql-operations'
import { DEFAULT_MINS_SAVED_PER_CHANGESET } from '../../../site-admin/analytics/AnalyticsBatchChangesPage'
import { ChangesetStatusClosed, ChangesetStatusOpen } from '../detail/changesets/ChangesetStatusCell'

import { GLOBAL_CHANGESETS_STATS } from './backend'

import styles from './BatchChangeStatsBar.module.scss'

interface BatchChangeStatsBarProps {
    className?: string
}

export const BatchChangeStatsBar: React.FunctionComponent<React.PropsWithChildren<BatchChangeStatsBarProps>> = () => {
    const { data, error, loading } = useQuery<GlobalChangesetsStatsResult, GlobalChangesetsStatsVariables>(
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

    const numMinPerItemSaved: number =
        parseInt(localStorage.getItem('minPerItemSaved') || '0', 10) || DEFAULT_MINS_SAVED_PER_CHANGESET

    return (
        <div className={styles.statsBar}>
            <div className={styles.leftSide}>
                <div className="pr-4">
                    <span className="font-weight-bold">{data?.batchChanges.totalCount}</span>
                    <br />
                    <span>Batch changes</span>
                </div>
                <div className="pr-4">
                    <span className="font-weight-bold">{data?.globalChangesetsStats.merged}</span>
                    <br />
                    <span>Merged</span>
                </div>
                <div className="pr-4">
                    <span className="font-weight-bold">
                        {data ? (data?.globalChangesetsStats.merged * numMinPerItemSaved) / 60 : '--'}
                    </span>
                    <br />
                    <span>
                        Hours Saved
                        <Tooltip content="Based on multiplier per changeset defined by site admin">
                            <Icon aria-label="More info" svgPath={mdiInformationOutline} className="ml-1" />
                        </Tooltip>
                    </span>
                </div>
            </div>
            <div className={styles.rightSide}>
                <div className="pr-4 text-center">
                    <ChangesetStatusOpen />
                    <span className="font-weight-bold">{data?.globalChangesetsStats.open}</span> <span>Open</span>
                </div>
                <div className="text-center">
                    <ChangesetStatusClosed />
                    <span className="font-weight-bold">{data?.globalChangesetsStats.closed}</span> <span>Closed</span>
                </div>
            </div>
        </div>
    )
}
