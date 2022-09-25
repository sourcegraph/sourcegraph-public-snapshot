import React from 'react'

import { useQuery } from '@sourcegraph/http-client'

import { ChangesetStatisticsResult, ChangesetStatisticsVariables } from '../../../graphql-operations'

import { CHANGESET_STATISTICS } from './backend'

import styles from './BatchChangeStatsBar.module.scss'

interface BatchChangeStatsBarProps {
    className?: string
}

export const BatchChangeStatsBar: React.FunctionComponent<React.PropsWithChildren<BatchChangeStatsBarProps>> = () => {
    const { data } = useQuery<ChangesetStatisticsResult, ChangesetStatisticsVariables>(CHANGESET_STATISTICS, {})

    console.log(data)

    return (
        <div className={styles.statsBar}>
            <div className={styles.leftSide}>
                <div className={styles.statItem}>
                    <span>{data?.batchChanges.totalCount}</span>
                    <br />
                    <span>Batch Changes</span>
                </div>
                <div className={styles.statItem}>
                    <span>{data?.merged.totalCount}</span>
                    <br />
                    <span>Merged</span>
                </div>
                <div className={styles.statItem}>
                    <span>27.7</span>
                    <br />
                    <span>Hours Saved</span>
                </div>
            </div>
            <div className={styles.rightSide}>
                <div className={styles.statItem}>
                    <span>image</span>
                    <br />
                    <span>{data?.opened.totalCount} Open</span>
                </div>
                <div>
                    <span>image</span>
                    <br />
                    <span>{data?.closed.totalCount} Closed</span>
                </div>
            </div>
        </div>
    )
}
