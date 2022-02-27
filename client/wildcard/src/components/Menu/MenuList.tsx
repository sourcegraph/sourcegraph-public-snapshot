import {
    MenuListProps as ReachMenuListProps,
    MenuPopover as ReachMenuPopover,
    MenuItems,
    MenuItemsProps,
} from '@reach/menu-button'
import classNames from 'classnames';
import React from 'react'

import { ForwardReferenceComponent } from '../../types'
import { createRectangle, PopoverContent, Position } from '../Popover'

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

const MENU_LIST_PADDING = createRectangle(0, 0, 2, 2)

export interface PopoverProps extends MenuItemsProps {
    popoverContentPosition?: Position
}

const Popover = React.forwardRef(({ popoverContentPosition, ...props }, reference) => (
    <PopoverContent
        {...props}
        ref={reference}
        position={popoverContentPosition}
        focusLocked={false}
        targetPadding={MENU_LIST_PADDING}
        as={MenuItems}
        className={classNames('py-1', props.className)}
    />
)) as ForwardReferenceComponent<'div', PopoverProps>
