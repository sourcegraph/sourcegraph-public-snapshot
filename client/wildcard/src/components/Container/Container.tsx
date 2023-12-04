import { forwardRef } from 'react'

import classNames from 'classnames'

import type { ForwardReferenceComponent } from '../../types'

import styles from './Container.module.scss'

/** A container wrapper. Used for grouping content together. */
export const Container = forwardRef((props, ref) => {
    const { as: Comp = 'div', className, ...attributes } = props

    return <Comp className={classNames(styles.container, className)} {...attributes} />
}) as ForwardReferenceComponent<'div'>
