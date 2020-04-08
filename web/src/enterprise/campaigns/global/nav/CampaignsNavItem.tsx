import React from 'react'
import { LinkWithIconOnlyTooltip } from '../../../../components/LinkWithIconOnlyTooltip'
import { CampaignsIcon } from '../../icons'

interface Props {
    className?: string
}

/**
 * An item in {@link GlobalNavbar} that links to the campaigns area.
 */
export const CampaignsNavItem: React.FunctionComponent<Props> = ({ className = '' }) => (
    <LinkWithIconOnlyTooltip
        to="/campaigns"
        text="Campaigns"
        icon={CampaignsIcon}
        className={`nav-link btn btn-link px-1 text-decoration-none e2e-campaign-nav-entry ${className}`}
        activeClassName="active"
    />
)
