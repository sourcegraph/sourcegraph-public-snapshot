import type { FC } from 'react'

import { Badge, type BadgeVariantType } from '@sourcegraph/wildcard'

import { type SearchJobNode, SearchJobState } from '../../../graphql-operations'

import styles from './SearchJobBadge.module.scss'

interface SearchJobBadgeProps {
    job: SearchJobNode
    withProgress?: boolean
}

export const SearchJobBadge: FC<SearchJobBadgeProps> = props => {
    const { job, withProgress } = props

    if (withProgress && job.state === SearchJobState.PROCESSING) {
        const totalRepo = job.repoStats.total
        const totalProcessedRepos = job.repoStats.completed

        return (
            <div className={styles.progress}>
                <div
                    // eslint-disable-next-line react/forbid-dom-props
                    style={{ width: `${100 * (totalProcessedRepos / totalRepo)}%` }}
                    className={styles.progressBar}
                />
            </div>
        )
    }

    return <Badge variant={getBadgeVariant(job.state)}>{job.state.toString()}</Badge>
}

const getBadgeVariant = (jobStatus: SearchJobState): BadgeVariantType | undefined => {
    switch (jobStatus) {
        case SearchJobState.COMPLETED: {
            return 'success'
        }
        case SearchJobState.QUEUED: {
            return 'secondary'
        }
        case SearchJobState.ERRORED: {
            return 'warning'
        }
        case SearchJobState.FAILED: {
            return 'danger'
        }
        case SearchJobState.PROCESSING: {
            return 'primary'
        }

        default: {
            return
        }
    }
}
