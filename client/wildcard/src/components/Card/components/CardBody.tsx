import React from 'react'

import classNames from 'classnames'

import type { ForwardReferenceComponent } from '../../..'

import styles from './CardBody.module.scss'

interface CardBodyProps {}

export const CardBody = React.forwardRef(({ as: Component = 'div', children, className, ...attributes }, reference) => (
    <Component ref={reference} className={classNames(styles.cardBody, className)} {...attributes}>
        {children}
    </Component>
)) as ForwardReferenceComponent<'div', CardBodyProps>
