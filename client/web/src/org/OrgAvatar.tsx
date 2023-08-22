import * as React from 'react'

import classNames from 'classnames'

import styles from './OrgAvatar.module.scss'

export interface OrgAvatarProps {
    /** The organization's name. */
    org: string

    light?: boolean

    size?: 'sm' | 'md' | 'lg'

    className?: string
}

const avatarSizeClasses: Record<NonNullable<OrgAvatarProps['size']>, string> = {
    sm: styles.orgAvatarSm,
    md: styles.orgAvatarMd,
    lg: styles.orgAvatarLg,
}

/**
 * OrgAvatar displays the avatar of an organization.
 */
export const OrgAvatar: React.FunctionComponent<React.PropsWithChildren<OrgAvatarProps>> = ({
    org,
    size = 'md',
    className = '',
    light,
}) => (
    <div
        className={classNames(styles.orgAvatar, avatarSizeClasses[size], className, {
            [styles.orgAvatarLight]: light,
        })}
    >
        {org.slice(0, 2).toUpperCase()}
    </div>
)
