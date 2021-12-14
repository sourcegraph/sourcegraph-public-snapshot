import { MenuItem as ReachMenuItem, MenuItemProps as ReachMenuItemProps } from '@reach/menu-button'
import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'

export type MenuItemProps = ReachMenuItemProps

export const MenuItem = React.forwardRef(({ children, className, ...props }, reference) => (
    <ReachMenuItem ref={reference} {...props} className={classNames('dropdown-item', className)}>
        {children}
    </ReachMenuItem>
)) as ForwardReferenceComponent<'div', MenuItemProps>
