import React from 'react'

import classNames from 'classnames'

import type { ForwardReferenceComponent } from '../../types'
import { Heading } from '../Typography'

import styles from './MenuHeader.module.scss'

export type MenuHeadingType = 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6'

/**
 * A simple styled header component that can be used to
 * label sections of a `<Menu />` component.
 */
export const MenuHeader = React.forwardRef(({ children, as: headerElement = 'h6', className, ...props }, reference) => (
    <Heading as={headerElement} ref={reference} {...props} className={classNames(styles.dropdownHeader, className)}>
        {children}
    </Heading>
)) as ForwardReferenceComponent<MenuHeadingType>
