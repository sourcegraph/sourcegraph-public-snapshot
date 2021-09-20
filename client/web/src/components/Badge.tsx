import classnames from 'classnames'
import React from 'react'

import styles from './Badge.module.scss'

export type BadgeStatus = 'beta' | 'prototype' | 'experimental' | 'wip' | 'new'

const statusStyleMapping: Record<BadgeStatus, string> = {
    prototype: 'badge-warning',
    wip: 'badge-warning',
    experimental: 'badge-info',
    beta: 'badge-info',
    new: 'badge-info',
}

export interface BadgeProps {
    status: BadgeStatus
    tooltip?: string
    className?: string
}

export const Badge: React.FunctionComponent<BadgeProps> = props => {
    const { className, status, tooltip } = props

    return (
        <span
            data-tooltip={tooltip}
            className={classnames(
                'badge',
                styles.badgeCapitalized,
                'd-inline-flex',
                'align-items-center',
                statusStyleMapping[status],
                className
            )}
        >
            {status}
        </span>
    )
}
