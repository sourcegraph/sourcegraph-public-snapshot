import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignsIcon } from '../icons'

interface Props {
    campaign: Pick<GQL.ICampaign, 'closedAt'>
    className?: string
}

export const CampaignStateBadge: React.FunctionComponent<Props> = ({ campaign, className = '' }) => {
    const isClosed = Boolean(campaign.closedAt)
    const badgeClassName = isClosed ? 'badge-danger' : 'badge-success'
    const badgeLabel = isClosed ? 'Closed' : 'Open'
    return (
        <span
            className={`badge ${badgeClassName} d-inline-flex align-items-center ${className} pr-2`}
            style={{ fontSize: 'unset' }}
        >
            <CampaignsIcon className="icon-inline mr-1" /> {badgeLabel}
        </span>
    )
}
