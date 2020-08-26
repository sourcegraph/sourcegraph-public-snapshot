import React from 'react'
import { CampaignFields } from '../../../graphql-operations'
import { Link } from '../../../../../shared/src/components/Link'
import { Timestamp } from '../../../components/time/Timestamp'

interface Props extends Pick<CampaignFields, 'createdAt' | 'initialApplier' | 'lastAppliedAt' | 'lastApplier'> {
    className?: string
}

/**
 * The created / updated at byline to the campaign header.
 */
export const CampaignInfoByline: React.FunctionComponent<Props> = ({
    className,
    createdAt,
    initialApplier,
    lastAppliedAt,
    lastApplier,
}) => (
    <div className={className}>
        <span className="text-muted">
            Created <Timestamp date={createdAt} /> by{' '}
            {initialApplier && <Link to={initialApplier.url}>{initialApplier.username}</Link>}
            {!initialApplier && <strong>deleted user</strong>}
        </span>
        <span className="mx-2 text-muted">|</span>
        <span className="text-muted">
            Updated <Timestamp date={lastAppliedAt} /> by{' '}
            {lastApplier && <Link to={lastApplier.url}>{lastApplier.username}</Link>}
            {!lastApplier && <strong>deleted user</strong>}
        </span>
    </div>
)
