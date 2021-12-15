import { MenuItems as ReachMenuItems, MenuItemsProps as ReachMenuItemsProps } from '@reach/menu-button'
import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'

export type MenuItemsProps = ReachMenuItemsProps

/**
 * A styled wrapper for `<MenuItem />` components that can be used
 * to give more control over the structure of the menu.
 *
 * @see — Docs https://reach.tech/menu-button#menuitems
 */
export const MenuItems = React.forwardRef(({ children, className, ...props }, reference) => (
    <ReachMenuItems ref={reference} {...props} className={classNames('d-block dropdown-menu', className)}>
        {children}
    </ReachMenuItems>
)) as ForwardReferenceComponent<'div', MenuItemsProps>
