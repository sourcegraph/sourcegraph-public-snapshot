import React, { FC, useMemo } from 'react'
import { Link } from 'react-router-dom'

import { encodeURIPathComponent } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { queryRepoChangesetsStats as _queryRepoChangesetsStats } from './backend'
import { BatchChangesIcon } from './icons'

interface RepoBatchChangesButtonProps {
    className?: string
    repoName: string
    /** For testing only. */
    queryRepoChangesetsStats?: typeof _queryRepoChangesetsStats
}

export const RepoBatchChangesButton: FC<RepoBatchChangesButtonProps> = ({
    className,
    repoName,
    queryRepoChangesetsStats = _queryRepoChangesetsStats,
}) => {
    const stats = useObservable(
        useMemo(() => queryRepoChangesetsStats({ name: repoName }), [queryRepoChangesetsStats, repoName])
    )

    if (!stats) {
        return null
    }

    const { open, merged } = stats.changesetsStats

    return (
        <Link className={className} to={`/${encodeURIPathComponent(repoName)}/-/batch-changes`}>
            <BatchChangesIcon className="icon-inline" /> Batch Changes
            {open > 0 && (
                <span
                    className="d-inline-block badge badge-success batch-change-badge ml-2"
                    data-tooltip={`${open} open batch changesets`}
                >
                    {open}
                </span>
            )}
            {merged > 0 && (
                <span
                    className="d-inline-block badge badge-merged batch-change-badge ml-2"
                    data-tooltip={`${merged} merged batch changesets`}
                >
                    {merged}
                </span>
            )}
        </Link>
    )
}
