import React from 'react'
import { Link } from 'react-router-dom'
import { DropdownItem, DropdownMenu } from 'reactstrap'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { useCampaigns } from '../../list/useCampaigns'

interface Props {
    className?: string
}

const LOADING: 'loading' = 'loading'

/**
 * A dropdown menu with a list of navigation links related to campaigns.
 */
export const CampaignsNavItemDropdownMenu: React.FunctionComponent<Props> = ({ className = '' }) => {
    const campaigns = useCampaigns()
    const MAX_CAMPAIGNS = 5
    return (
        <DropdownMenu className={className} style={{ maxWidth: '12rem' }}>
            {campaigns === LOADING ? (
                <DropdownItem header={true} className="py-1">
                    Loading campaigns...
                </DropdownItem>
            ) : isErrorLike(campaigns) ? (
                <DropdownItem header={true} className="py-1">
                    Error loading campaigns
                </DropdownItem>
            ) : (
                <>
                    <DropdownItem header={true} className="py-1">
                        Recent campaigns
                    </DropdownItem>
                    {campaigns.nodes.slice(0, MAX_CAMPAIGNS).map(campaign => (
                        <Link key={campaign.id} to={campaign.url} className="dropdown-item text-truncate">
                            {campaign.name}
                        </Link>
                    ))}
                    <DropdownItem divider={true} />
                </>
            )}
            <Link to="/campaigns" className="dropdown-item">
                All campaigns
            </Link>
        </DropdownMenu>
    )
}
