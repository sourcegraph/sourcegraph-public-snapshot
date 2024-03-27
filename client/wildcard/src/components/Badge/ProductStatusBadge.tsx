import React from 'react'

import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'

import type { ForwardReferenceComponent } from '../../types'

import { type BadgeProps, Badge } from './Badge'
import type { BADGE_VARIANTS, PRODUCT_STATUSES } from './constants'

import styles from './ProductStatusBadge.module.scss'

export type ProductStatusType = typeof PRODUCT_STATUSES[number]

/**
 * Product statuses mapped to Badge style variants
 */
const STATUS_VARIANT_MAPPING: Record<ProductStatusType, typeof BADGE_VARIANTS[number]> = {
    wip: 'warning',
    experimental: 'warning',
    beta: 'info',
    'private-beta': 'info',
    new: 'info',
}

type Extends<T, U extends T> = U
export type ProductStatusLinked = Extends<ProductStatusType, 'beta' | 'experimental' | 'private-beta'>

/**
 * Map badge status to a relevant docs page describing that product status
 */
const STATUS_LINK_MAPPING: Record<ProductStatusLinked, string> = {
    experimental: 'https://sourcegraph.com/docs/admin/beta_and_experimental_features#experimental-features',
    beta: 'https://sourcegraph.com/docs/admin/beta_and_experimental_features#beta-features',
    'private-beta': 'https://sourcegraph.com/docs/admin/beta_and_experimental_features#beta-features',
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
export const ProductStatusBadge = React.forwardRef(function ProductStatusBadge(props, reference) {
    const variant = STATUS_VARIANT_MAPPING[props.status]
    const className = classNames(styles.productStatusBadge, props.className)
    const label =
        props.status === 'beta'
            ? 'This feature is currently in beta'
            : props.status === 'experimental'
            ? 'This feature is experimental'
            : props.status === 'wip'
            ? 'This feature is a work in progress'
            : props.status === 'new'
            ? 'This feature is new'
            : props.status === 'private-beta'
            ? 'This feature is in private beta'
            : ''
    const status = props.status === 'private-beta' ? 'private beta' : props.status

    if ('linkToDocs' in props) {
        return (
            <>
                <VisuallyHidden>{label}</VisuallyHidden>
                <Badge
                    ref={reference}
                    href={STATUS_LINK_MAPPING[props.status as ProductStatusLinked]}
                    variant={variant}
                    className={className}
                    aria-hidden={true}
                >
                    {status}
                </Badge>
            </>
        )
    }

    return (
        <>
            <VisuallyHidden>{`This feature is currently in ${props.status}`}</VisuallyHidden>
            <Badge ref={reference} {...props} variant={variant} className={className} aria-hidden={true}>
                {status}
            </Badge>
        </>
    )
}) as ForwardReferenceComponent<'span', ProductStatusBadgeProps>
