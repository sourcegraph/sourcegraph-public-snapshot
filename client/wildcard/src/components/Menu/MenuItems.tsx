import { MenuItems as ReachMenuItems, MenuItemsProps as ReachMenuItemsProps } from '@reach/menu-button'
import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'

export type MenuItemsProps = ReachMenuItemsProps

export const MenuItems = React.forwardRef(({ children, className, ...props }, reference) => (
    <ReachMenuItems ref={reference} {...props} className={classNames(className, 'dropdown-menu')}>
        {children}
    </ReachMenuItems>
)) as ForwardReferenceComponent<'div', MenuItemsProps>
