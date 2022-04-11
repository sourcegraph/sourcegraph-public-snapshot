import React from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../types'

export interface MenuTextProps {}

/**
 * A simple styled wrapper component that can be used
 * in and/or outside the Menu context.
 */
export const MenuText = React.forwardRef(({ children, as: Component = 'div', className, ...props }, reference) => (
    <Component role="menuitem" ref={reference} {...props} className={classNames('dropdown-item', className)}>
        {children}
    </Component>
)) as ForwardReferenceComponent<'div', MenuTextProps>
