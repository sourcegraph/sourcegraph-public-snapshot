import React from 'react'

import { MenuPopover as ReachMenuPopover, MenuItems } from '@reach/menu-button'
import { Position as ReachPopoverPosition } from '@reach/popover'
import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../types'
import { createRectangle, PopoverContent, PopoverContentProps, Position } from '../Popover'

const DEFAULT_MENU_LIST_PADDING = createRectangle(0, 0, 2, 2)

export interface MenuListProps extends PopoverProps {
    position?: Position
}

export const MenuList = React.forwardRef((props, reference) => {
    const { children, position = Position.bottomStart, targetPadding = DEFAULT_MENU_LIST_PADDING, ...rest } = props

    return (
        <ReachMenuPopover
            {...rest}
            as={Popover}
            ref={reference}
            portal={false}
            targetPadding={targetPadding}
            // Since both ReachMenuPopover and Popover components have position prop
            // we have to cast one to other in order to avoid type collision problem
            position={(position as unknown) as ReachPopoverPosition}
        >
            {children}
        </ReachMenuPopover>
    )
}) as ForwardReferenceComponent<'div', MenuListProps>

export interface PopoverProps extends PopoverContentProps {}

const Popover = React.forwardRef(({ ...props }, reference) => (
    <PopoverContent
        {...props}
        as={MenuItems}
        ref={reference}
        focusLocked={false}
        className={classNames('py-1', props.className)}
    />
)) as ForwardReferenceComponent<'div', PopoverProps>
