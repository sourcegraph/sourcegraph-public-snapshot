import React from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../types'

/**
 * A simple styled divider that can be used within a
 * `<Menu />` component to separate menu items.
 */
export const MenuDivider = React.forwardRef(({ children, className, ...props }, reference) => (
    <div ref={reference} {...props} className={classNames('dropdown-divider', className)} />
)) as ForwardReferenceComponent<'div'>
