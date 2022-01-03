import classNames from 'classnames'
import React, { FC, useMemo } from 'react'
import { Link } from 'react-router-dom'

import { encodeURIPathComponent } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Badge } from '@sourcegraph/wildcard'

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
        <Link
            className={classNames('btn btn-outline-secondary', className)}
            to={`/${encodeURIPathComponent(repoName)}/-/batch-changes`}
        >
            <BatchChangesIcon className="icon-inline" /> Batch Changes
            {open > 0 && (
                <Badge
                    tooltip={`${open} open batch changesets`}
                    variant="success"
                    className="d-inline-block batch-change-badge ml-2"
                >
                    {open}
                </Badge>
            )}
            {merged > 0 && (
                <Badge
                    tooltip={`${merged} merged batch changesets`}
                    variant="merged"
                    className="d-inline-block batch-change-badge ml-2"
                >
                    {merged}
                </Badge>
            )}
        </Link>
    )
}
