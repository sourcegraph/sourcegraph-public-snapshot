import React from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../types'

export type MenuHeadingType = 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6'

/**
 * A simple styled header component that can be used to
 * label sections of a `<Menu />` component.
 */

export const MenuHeader = React.forwardRef(({ children, as: Component = 'h6', className, ...props }, reference) => (
    <Component ref={reference} {...props} className={classNames('dropdown-header', className)}>
        {children}
    </Component>
)) as ForwardReferenceComponent<MenuHeadingType>
