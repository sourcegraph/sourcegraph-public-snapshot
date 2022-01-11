import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../..'

import styles from './CardText.module.scss'

interface CardTextProps {
    /**
     * Used to change the element that is rendered.
     */
    as?: React.ElementType
}

export const CardText = React.forwardRef(({ as: Component = 'p', children, className, ...attributes }, reference) => (
    <Component ref={reference} className={classNames(styles.cardText, className)} {...attributes}>
        {children}
    </Component>
)) as ForwardReferenceComponent<'p', CardTextProps>
