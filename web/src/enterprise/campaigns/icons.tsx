import React from 'react'
import ImageAutoAdjustIcon from 'mdi-react/ImageAutoAdjustIcon'

/**
 * The icon to use everywhere to represent a campaign. If the icon's left side needs to be flush
 * with the left edge, use {@link CampaignsIconFlushEdges} instead. (Same goes for any other side,
 * but flush left is the most common.)
 */
export const CampaignsIcon = ImageAutoAdjustIcon

/**
 * The same icon as {@link CampaignsIcon}, except the icon has no padding. This is important when,
 * for example, the icon's left edge needs to be flush with the left edges of other content
 * displayed above and below it.
 *
 * The only difference is in the following attribute: `<svg viewBox="3 3 18 18">`.
 */
const CampaignsIconFlushEdges: React.FunctionComponent<{ className?: string }> = React.memo(({ className = '' }) => (
    <svg fill="currentColor" width={24} height={24} className={`mdi-icon ${className}`} viewBox="3 3 18 18">
        <path d="M19 10V19H5V5H14V3H5C3.92 3 3 3.9 3 5V19C3 20.1 3.92 21 5 21H19C20.12 21 21 20.1 21 19V10H19M17 10L17.94 7.94L20 7L17.94 6.06L17 4L16.06 6.06L14 7L16.06 7.94L17 10M13.25 10.75L12 8L10.75 10.75L8 12L10.75 13.25L12 16L13.25 13.25L16 12L13.25 10.75Z" />
    </svg>
))

export const CampaignsIconWithBetaBadge: React.FunctionComponent<{ className?: string }> = ({ className = '' }) => (
    <>
        <CampaignsIconFlushEdges className={className} />{' '}
        <sup>
            <span className="badge badge-merged text-uppercase">Beta</span>
        </sup>
    </>
)
