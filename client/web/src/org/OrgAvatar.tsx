import classNames from 'classnames'
import * as React from 'react'

import styles from './OrgAvatar.module.scss'

export interface OrgAvatarProps {
    /** The organization's name. */
    org: string

    size?: 'md' | 'lg'

    className?: string
}

const avatarSizeClasses: Record<NonNullable<OrgAvatarProps['size']>, string> = {
    md: styles.orgAvatarMd,
    lg: styles.orgAvatarLg,
}

/**
 * OrgAvatar displays the avatar of an organization.
 */
export const OrgAvatar: React.FunctionComponent<OrgAvatarProps> = ({ org, size = 'md', className = '' }) => (
    <div className={classNames(styles.orgAvatar, avatarSizeClasses[size], className)}>
        {org.slice(0, 2).toUpperCase()}
    </div>
)
