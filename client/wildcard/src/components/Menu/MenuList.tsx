import {
    MenuListProps as ReachMenuListProps,
    MenuPopover as ReachMenuPopover,
    MenuItems,
    MenuItemsProps,
} from '@reach/menu-button'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'
import { PopoverContent, Position } from '../Popover'

export interface MenuListProps extends Omit<ReachMenuListProps, 'position' | 'portal'> {
    position?: Position
}

export const MenuList = React.forwardRef((props, reference) => {
    const { children, position, ...rest } = props

    return (
        <ReachMenuPopover {...rest} ref={reference} popoverContentPosition={position} portal={false} as={Popover}>
            {children}
        </ReachMenuPopover>
    )
}) as ForwardReferenceComponent<'div', MenuListProps>

export interface PopoverProps extends MenuItemsProps {
    popoverContentPosition?: Position
}

const Popover = React.forwardRef((props, reference) => (
    <PopoverContent
        {...props}
        ref={reference}
        position={props.popoverContentPosition}
        focusLocked={false}
        as={MenuItems}
    />
)) as ForwardReferenceComponent<'div', PopoverProps>
