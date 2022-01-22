import { MenuListProps, MenuList as ReachMenuList, MenuItems, MenuItemsProps } from '@reach/menu-button'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'
import { PopoverContent } from '../Popover'

export const MenuList = React.forwardRef(({ portal, children, ...props }, reference) => (
    <ReachMenuList {...props} ref={reference} portal={false} as={Popover}>
        {children}
    </ReachMenuList>
)) as ForwardReferenceComponent<'div', MenuListProps>

const Popover = React.forwardRef((props, reference) => (
    <PopoverContent {...props} ref={reference} focusLocked={false} as={MenuItems} />
)) as ForwardReferenceComponent<'div', MenuItemsProps>
