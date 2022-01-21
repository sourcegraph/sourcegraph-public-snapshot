import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent, useWildcardTheme } from '../..'

import styles from './Card.module.scss'

interface CardProps {
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
export const Card = React.forwardRef(({ children, className, as: Component = 'div', ...attributes }, reference) => {
    const { isBranded } = useWildcardTheme()

    return (
        <Component className={classNames(isBranded && styles.card, className)} ref={reference} {...attributes}>
            {children}
        </Component>
    )
}) as ForwardReferenceComponent<'div', CardProps>
