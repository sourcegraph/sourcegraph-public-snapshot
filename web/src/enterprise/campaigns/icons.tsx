import React from 'react'
import ImageAutoAdjustIcon from 'mdi-react/ImageAutoAdjustIcon'

/**
 * The icon to use everywhere to represent a campaign
 */
export const CampaignsIcon = ImageAutoAdjustIcon

export const CampaignsIconWithBetaBadge: React.FunctionComponent<{ className?: string }> = ({ className = '' }) => (
    <>
        <CampaignsIcon className={className} />{' '}
        <sup>
            <span className="badge badge-merged text-uppercase">Beta</span>
        </sup>
    </>
)
