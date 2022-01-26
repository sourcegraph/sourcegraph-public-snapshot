import classNames from 'classnames'
import React from 'react'

import { useWildcardTheme } from '../../hooks/useWildcardTheme'
import { Link } from '../Link'

import styles from './Badge.module.scss'
import { BADGE_VARIANTS } from './constants'

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
    /**
     * The variant style of the badge.
     */
    variant?: typeof BADGE_VARIANTS[number]
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
     */
    href?: string
    /**
     * Used to change the element that is rendered.
     */
    as?: React.ElementType
    className?: string
}

/**
 * An abstract UI component which renders a small "badge" with specific styles to help annotate content.
 */
export const Badge: React.FunctionComponent<BadgeProps> = ({
    children,
    variant,
    small,
    pill,
    tooltip,
    className,
    href,
    as: Component = 'span',
    ...otherProps
}) => {
    const { isBranded } = useWildcardTheme()
    const brandedClassName =
        isBranded && classNames(styles.badge, variant && styles[variant], small && styles.sm, pill && styles.pill)

    const commonProps = {
        'data-tooltip': tooltip,
        className: classNames(brandedClassName, className),
        ...otherProps,
    }

    if (href) {
        return (
            <Link to={href} rel="noopener" target="_blank" {...commonProps}>
                {children}
            </Link>
        )
    }

    return <Component {...commonProps}>{children}</Component>
}
