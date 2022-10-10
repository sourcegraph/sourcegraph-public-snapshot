import React from 'react'

import { mdiInformationOutline } from '@mdi/js'

import { useQuery } from '@sourcegraph/http-client'
import { H3, Icon, LoadingSpinner, Tooltip } from '@sourcegraph/wildcard'

import { GlobalChangesetsStatsResult, GlobalChangesetsStatsVariables } from '../../../graphql-operations'
import { DEFAULT_MINS_SAVED_PER_CHANGESET } from '../../../site-admin/analytics/AnalyticsBatchChangesPage'
import { MIN_PER_ITEM_SAVED_KEY } from '../../../site-admin/analytics/components/TimeSavedCalculatorGroup'
import { ChangesetStatusClosed, ChangesetStatusOpen } from '../detail/changesets/ChangesetStatusCell'

import { GLOBAL_CHANGESETS_STATS } from './backend'

import styles from './BatchChangeStatsBar.module.scss'

interface BatchChangeStatsBarProps {
    className?: string
}

export const BatchChangeStatsBar: React.FunctionComponent<React.PropsWithChildren<BatchChangeStatsBarProps>> = () => {
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

    const numMinPerItemSaved: number =
        parseInt(localStorage.getItem(MIN_PER_ITEM_SAVED_KEY) || '0', 10) || DEFAULT_MINS_SAVED_PER_CHANGESET

    return (
        <div className={styles.statsBar}>
            <div className={styles.leftSide}>
                <div className="pr-4">
                    <H3 className="font-weight-bold mb-1">{data.batchChanges.totalCount}</H3>
                    <span>Batch changes</span>
                </div>
                <div className="pr-4">
                    <H3 className="font-weight-bold mb-1">{data.globalChangesetsStats.merged}</H3>
                    <span>Changesets merged</span>
                </div>
                <div className="pr-4">
                    <H3 className="font-weight-bold mb-1">
                        {(data.globalChangesetsStats.merged * numMinPerItemSaved) / 60}
                    </H3>
                    <span>
                        Hours saved
                        <Tooltip content="Based on multiplier per changeset defined by site admin">
                            <Icon aria-label="More info" svgPath={mdiInformationOutline} className="ml-1" />
                        </Tooltip>
                    </span>
                </div>
            </div>
            <div className={styles.rightSide}>
                <div className="pr-4 text-center text-muted">
                    <ChangesetStatusOpen />
                    <span>{data.globalChangesetsStats.open} open</span>
                </div>
                <div className="text-center text-muted">
                    <ChangesetStatusClosed />
                    <span>{data.globalChangesetsStats.closed} closed</span>
                </div>
            </div>
        </div>
    )
}
