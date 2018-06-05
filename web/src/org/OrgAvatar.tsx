import * as React from 'react'

/**
 * OrgAvatar displays the avatar of an organization.
 */
export const OrgAvatar: React.SFC<{
    /** The organization's name. */
    org: string

    className?: string
}> = ({ org, className = '' }) => <div className={`org-avatar ${className}`}>{org.substr(0, 2).toUpperCase()}</div>
