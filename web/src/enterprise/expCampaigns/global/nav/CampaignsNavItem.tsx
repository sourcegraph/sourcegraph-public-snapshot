import React from 'react'
import { LinkWithIconOnlyTooltip } from '../../../../components/LinkWithIconOnlyTooltip'
import { CampaignsIcon } from '../../icons'

interface Props {
    className?: string
}

/**
 * An item in {@link GlobalNavbar} that links to the changesets area.
 */
export const CampaignsNavItem: React.FunctionComponent<Props> = ({ className = '' }) => (
    <LinkWithIconOnlyTooltip
        to="/exp/campaigns"
        text="Campaigns"
        icon={CampaignsIcon}
        className={`nav-link btn btn-link px-3 text-decoration-none ${className}`}
    />
)
