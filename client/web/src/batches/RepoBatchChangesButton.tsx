import { FC } from 'react'

import { encodeURIPathComponent, pluralize } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { Badge, Button, Link, Icon } from '@sourcegraph/wildcard'

import { RepoChangesetsStatsResult, RepoChangesetsStatsVariables } from '../graphql-operations'

import { REPO_CHANGESETS_STATS } from './backend'
import { BatchChangesIcon } from './icons'

interface RepoBatchChangesButtonProps {
    className?: string
    repoName: string
}

export const RepoBatchChangesButton: FC<React.PropsWithChildren<RepoBatchChangesButtonProps>> = ({
    className,
    repoName,
}) => {
    const { data } = useQuery<RepoChangesetsStatsResult, RepoChangesetsStatsVariables>(REPO_CHANGESETS_STATS, {
        variables: { name: repoName },
        fetchPolicy: 'cache-first',
    })

    if (!data || !data.repository) {
        return null
    }

    const { open, merged } = data.repository.changesetsStats

    return (
        <Button
            className={className}
            to={`/${encodeURIPathComponent(repoName)}/-/batch-changes`}
            variant="secondary"
            outline={true}
            as={Link}
        >
            <Icon as={BatchChangesIcon} aria-hidden={true} /> Batch Changes
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
