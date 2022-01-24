import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../..'

import styles from './CardFooter.module.scss'

interface CardFooterProps {
    /**
     * Used to change the element that is rendered.
     */
    as?: React.ElementType
}

export const CardFooter = React.forwardRef(
    ({ as: Component = 'div', children, className, ...attributes }, reference) => (
        <Component ref={reference} className={classNames(styles.cardFooter, className)} {...attributes}>
            {children}
        </Component>
    )
) as ForwardReferenceComponent<'div', CardFooterProps>
