import React from 'react'

import classNames from 'classnames'

import { BadgeProps, Badge } from './Badge'
import { BADGE_VARIANTS, PRODUCT_STATUSES } from './constants'

import styles from './ProductStatusBadge.module.scss'

export type ProductStatusType = typeof PRODUCT_STATUSES[number]

/**
 * Product statuses mapped to Badge style variants
 */
const STATUS_VARIANT_MAPPING: Record<ProductStatusType, typeof BADGE_VARIANTS[number]> = {
    prototype: 'warning',
    wip: 'warning',
    experimental: 'warning',
    beta: 'info',
    new: 'info',
}

type Extends<T, U extends T> = U
export type ProductStatusLinked = Extends<ProductStatusType, 'beta' | 'experimental'>

/**
 * Map badge status to a relevant docs page describing that product status
 */
const STATUS_LINK_MAPPING: Record<ProductStatusLinked, string> = {
    experimental: 'https://docs.sourcegraph.com/admin/beta_and_experimental_features#experimental-features',
    beta: 'https://docs.sourcegraph.com/admin/beta_and_experimental_features#beta-features',
} as const

/**
 * Badge props without custom style configuration.
 * We handle this ourselves based on the current badge `status`.
 */
type BaseBadgeProps = Omit<BadgeProps, 'variant' | 'size'>

export interface BaseProductStatusBadgeProps extends BaseBadgeProps {
    status: ProductStatusType
}
export interface PossibleLinkedProductStatusBadge extends BaseBadgeProps {
    status: ProductStatusLinked
    /**
     * Whether this badge should link to the relevant documentation page for this status
     */
    linkToDocs?: boolean
}
export type ProductStatusBadgeProps = BaseProductStatusBadgeProps | PossibleLinkedProductStatusBadge

/**
 * A specific Badge component wrapper to describe a product status.
 * Can also be used to link to the relevant docs page for that status.
 */
export const ProductStatusBadge: React.FunctionComponent<React.PropsWithChildren<ProductStatusBadgeProps>> = props => {
    const variant = STATUS_VARIANT_MAPPING[props.status]
    const className = classNames(styles.productStatusBadge, props.className)

    if ('linkToDocs' in props) {
        return (
            <Badge href={STATUS_LINK_MAPPING[props.status]} variant={variant} className={className}>
                {props.status}
            </Badge>
        )
    }

    return (
        <Badge {...props} variant={variant} className={className}>
            {props.status}
        </Badge>
    )
}
