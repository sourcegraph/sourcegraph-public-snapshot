import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../..'

import styles from './CardTitle.module.scss'

interface CardTitleProps {
    /**
     * Used to change the element that is rendered.
     */
    as?: React.ElementType
}

export const CardTitle = React.forwardRef(({ as: Component = 'h3', children, className, ...attributes }, reference) => (
    <Component ref={reference} className={classNames(styles.cardTitle, className)} {...attributes}>
        {children}
    </Component>
)) as ForwardReferenceComponent<'h3', CardTitleProps>
