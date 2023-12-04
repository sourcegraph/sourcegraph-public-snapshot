import React from 'react'

import classNames from 'classnames'

import { type ForwardReferenceComponent, H3 } from '../../..'

import styles from './CardTitle.module.scss'

interface CardTitleProps {}

export const CardTitle = React.forwardRef(({ as: Component = H3, children, className, ...attributes }, reference) => (
    <Component ref={reference} className={classNames(styles.cardTitle, className)} {...attributes}>
        {children}
    </Component>
)) as ForwardReferenceComponent<'h3', CardTitleProps>
