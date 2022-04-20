import { forwardRef } from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../..'

import styles from './CardSubtitle.module.scss'

interface CardSubtitleProps {}

export const CardSubtitle = forwardRef(({ as: Component = 'div', children, className, ...attributes }, reference) => (
    <Component ref={reference} className={classNames(styles.cardSubtitle, className)} {...attributes}>
        {children}
    </Component>
)) as ForwardReferenceComponent<'div', CardSubtitleProps>
