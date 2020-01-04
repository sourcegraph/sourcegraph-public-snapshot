import React from 'react'
import classNames from 'classnames'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignsIcon } from '../icons'

interface Props {
    campaign: Pick<GQL.ICampaign, 'closedAt'>

    className?: string
}

/**
 * A badge that conveys the status of a campaign.
 */
export const CampaignStatusBadge: React.FunctionComponent<Props> = ({ campaign, className }) => (
    <span
        className={classNames(
            'campaign-status-badge badge',
            className,
            !campaign.closedAt ? 'badge-success' : 'badge-danger'
        )}
    >
        <CampaignsIcon className="icon-inline mr-1" /> {!campaign.closedAt ? 'Open' : 'Closed'}
    </span>
)
