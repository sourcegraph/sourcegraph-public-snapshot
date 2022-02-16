import {
    MenuListProps as ReachMenuListProps,
    MenuPopover as ReachMenuPopover,
    MenuItems,
    MenuItemsProps,
} from '@reach/menu-button'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'
import { PopoverContent, Position, Strategy } from '../Popover'

export interface MenuListProps extends Omit<ReachMenuListProps, 'position' | 'portal'> {
    position?: Position
    strategy?: Strategy
}

export const MenuList = React.forwardRef((props, reference) => {
    const { children, position, strategy, ...rest } = props

    return (
        <ReachMenuPopover
            {...rest}
            ref={reference}
            popoverContentPosition={position}
            popoverContentStrategy={strategy}
            portal={false}
            as={Popover}
        >
            {children}
        </ReachMenuPopover>
    )
}) as ForwardReferenceComponent<'div', MenuListProps>

export interface PopoverProps extends MenuItemsProps {
    popoverContentPosition?: Position
    popoverContentStrategy?: Strategy
}

const Popover = React.forwardRef(({ popoverContentPosition, popoverContentStrategy, ...props }, reference) => (
    <PopoverContent
        {...props}
        ref={reference}
        position={popoverContentPosition}
        strategy={popoverContentStrategy}
        focusLocked={false}
        as={MenuItems}
    />
)) as ForwardReferenceComponent<'div', PopoverProps>
