import classNames from 'classnames'
import React from 'react'

import styles from './Card.module.scss'
import { CARD_VARIANTS } from './constants'

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
    /**
     * Interactive variants, shows blue border on hover and focus
     */
    variant?: typeof CARD_VARIANTS[number]
    /**
     * Used to change the element that renders card content.
     * Useful if needing to provide interactive elements for the interactive-card variant.
     * Defaults to button on interactive, and div for non-interactive card variants.
     */
    as?: React.ElementType
}

/**
 * Card Element
 */
export const Card: React.FunctionComponent<CardProps> = ({
    children,
    className,
    variant = 'default',
    as: Component = variant === 'interactive' ? 'button' : 'div',
    ...attributes
}) => (
    <Component
        className={classNames(styles.card, className, variant === 'interactive' && styles.cardInteractive)}
        {...attributes}
    >
        {children}
    </Component>
)
