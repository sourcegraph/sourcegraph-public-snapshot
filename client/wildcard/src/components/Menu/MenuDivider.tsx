import React from 'react'

import classNames from 'classnames'

import type { ForwardReferenceComponent } from '../../types'

import styles from './MenuDivider.module.scss'

/**
 * A simple styled divider that can be used within a
 * `<Menu />` component to separate menu items.
 */
export const MenuDivider = React.forwardRef(({ children, className, ...props }, reference) => (
    <div ref={reference} {...props} className={classNames(styles.dropdownDivider, className)} />
)) as ForwardReferenceComponent<'div'>
