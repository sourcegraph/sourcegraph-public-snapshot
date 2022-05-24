import { FC, useMemo } from 'react'

import { encodeURIPathComponent, pluralize } from '@sourcegraph/common'
import { Badge, useObservable, Button, Link, Icon } from '@sourcegraph/wildcard'

import { queryRepoChangesetsStats as _queryRepoChangesetsStats } from './backend'
import { BatchChangesIcon } from './icons'

interface RepoBatchChangesButtonProps {
    className?: string
    repoName: string
    /** For testing only. */
    queryRepoChangesetsStats?: typeof _queryRepoChangesetsStats
}

export const RepoBatchChangesButton: FC<React.PropsWithChildren<RepoBatchChangesButtonProps>> = ({
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
        <Button
            className={className}
            to={`/${encodeURIPathComponent(repoName)}/-/batch-changes`}
            variant="secondary"
            outline={true}
            as={Link}
        >
            <Icon as={BatchChangesIcon} /> Batch Changes
            {open > 0 && (
                <Badge
                    tooltip={`${open} open ${pluralize('batch changeset', open)}`}
                    variant="success"
                    className="d-inline-block batch-change-badge ml-2"
                >
                    {open}
                </Badge>
            )}
            {merged > 0 && (
                <Badge
                    tooltip={`${merged} merged ${pluralize('batch changeset', merged)}`}
                    variant="merged"
                    className="d-inline-block batch-change-badge ml-2"
                >
                    {merged}
                </Badge>
            )}
        </Button>
    )
}
