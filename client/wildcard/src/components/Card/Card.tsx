import React from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent, useWildcardTheme } from '../..'

import styles from './Card.module.scss'

export interface CardProps {}

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
