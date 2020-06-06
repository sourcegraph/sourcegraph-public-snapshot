import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { pluralize } from '../../../../../../shared/src/util/strings'

interface Props {
    /** The latest changeset states. */
    changesetCounts: Pick<GQL.IChangesetCounts, 'total' | 'merged' | 'closed' | 'open'>

    className?: string
}

/**
 * A badge with the campaign's progress toward completion (as a percentage).
 */
export const CampaignProgressBadge: React.FunctionComponent<Props> = ({ changesetCounts, className = '' }) => {
    const completed = changesetCounts.merged + changesetCounts.closed
    return changesetCounts.total > 0 ? (
        <span className={className}>
            {Math.floor((completed / changesetCounts.total) * 100)}% complete · {changesetCounts.total}{' '}
            {pluralize('changeset', changesetCounts.total)} total
        </span>
    ) : null
}
