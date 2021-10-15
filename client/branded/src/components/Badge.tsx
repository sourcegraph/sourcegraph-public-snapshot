import classNames from 'classnames'
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

type Extends<T, U extends T> = U
type BadgeStatusLinked = Extends<BadgeStatus, 'beta' | 'experimental'>

const statusLinkMapping: Record<BadgeStatusLinked, string> = {
    experimental: 'https://docs.sourcegraph.com/admin/beta_and_experimental_features#experimental-features',
    beta: 'https://docs.sourcegraph.com/admin/beta_and_experimental_features#beta-features',
}

export interface BadgeProps {
    status: BadgeStatus
    tooltip?: string
    className?: string
    useLink?: boolean
}

export const Badge: React.FunctionComponent<BadgeProps> = props => {
    const { className, status, tooltip, useLink } = props

    const commonProps = {
        'data-tooltip': tooltip,
        className: classNames(
            'badge',
            styles.badgeCapitalized,
            'd-inline-flex',
            'align-items-center',
            statusStyleMapping[status],
            className
        ),
    }

    if (useLink && statusLinkMapping[status as BadgeStatusLinked]) {
        return (
            <a href={statusLinkMapping[status as BadgeStatusLinked]} rel="noopener" target="_blank" {...commonProps}>
                {status}
            </a>
        )
    }

    return <span {...commonProps}>{status}</span>
}
