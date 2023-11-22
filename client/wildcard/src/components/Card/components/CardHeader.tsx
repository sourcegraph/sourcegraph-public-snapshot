import React from 'react'

import classNames from 'classnames'

import type { ForwardReferenceComponent } from '../../..'

import styles from './CardHeader.module.scss'

interface CardHeaderProps {}

export const CardHeader = React.forwardRef(
    ({ as: Component = 'div', children, className, ...attributes }, reference) => (
        <Component ref={reference} className={classNames(styles.cardHeader, className)} {...attributes}>
            {children}
        </Component>
    )
) as ForwardReferenceComponent<'div', CardHeaderProps>
