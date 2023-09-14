import React from 'react'

import classNames from 'classnames'

import { useWildcardTheme } from '../../hooks/useWildcardTheme'
import type { ForwardReferenceComponent } from '../../types'
import { Link } from '../Link'
import { Tooltip } from '../Tooltip'

import type { BADGE_VARIANTS } from './constants'

import styles from './Badge.module.scss'

export type BadgeVariantType = typeof BADGE_VARIANTS[number]

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
    /**
     * The variant style of the badge.
     */
    variant?: BadgeVariantType
    /**
     * Allows modifying the size of the badge. Supports a smaller variant.
     */
    small?: boolean
    /**
     * Render the badge as a rounded pill
     */
    pill?: boolean
    /**
     * Additional text to display on hover
     */
    tooltip?: string
    /**
     * Used to render the badge as a link to a specific URL
     *
     * @deprecated Use `as` prop instead
     */
    href?: string
    className?: string
}

/**
 * An abstract UI component which renders a small "badge" with specific styles to help annotate content.
 */
export const Badge = React.forwardRef(function Badge(
    { children, variant, small, pill, tooltip, className, href, as: Component = 'span', ...otherProps },
    reference
) {
    const { isBranded } = useWildcardTheme()
    const brandedClassName =
        isBranded && classNames(styles.badge, variant && styles[variant], small && styles.sm, pill && styles.pill)

    const commonProps = {
        className: classNames(brandedClassName, className),
        ...otherProps,
    }

    if (href) {
        return (
            <Tooltip content={tooltip}>
                <Link to={href} rel="noopener" target="_blank" {...commonProps} ref={null}>
                    {children}
                </Link>
            </Tooltip>
        )
    }

    return (
        <Tooltip content={tooltip}>
            <Component {...commonProps} ref={reference}>
                {children}
            </Component>
        </Tooltip>
    )
}) as ForwardReferenceComponent<'span', BadgeProps>
