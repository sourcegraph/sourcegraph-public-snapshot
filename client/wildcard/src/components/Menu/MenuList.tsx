import { MenuListProps, MenuPopover as ReachMenuPopover, MenuItems, MenuItemsProps } from '@reach/menu-button'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'
import { PopoverContent } from '../Popover'

export const MenuList = React.forwardRef(({ portal, children, ...props }, reference) => (
    <ReachMenuPopover {...props} ref={reference} portal={false} as={Popover}>
        {children}
    </ReachMenuPopover>
)) as ForwardReferenceComponent<'div', MenuListProps>

const Popover = React.forwardRef((props, reference) => (
    <PopoverContent {...props} ref={reference} focusLocked={false} as={MenuItems} />
)) as ForwardReferenceComponent<'div', MenuItemsProps>
