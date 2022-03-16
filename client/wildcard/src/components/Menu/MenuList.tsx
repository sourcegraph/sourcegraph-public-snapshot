import React from 'react'

import { MenuPopover as ReachMenuPopover, MenuItems } from '@reach/menu-button'
import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../types'
import { createRectangle, PopoverContent, PopoverContentProps, Position } from '../Popover'

const DEFAULT_MENU_LIST_PADDING = createRectangle(0, 0, 2, 2)

export interface MenuListProps extends Omit<PopoverProps, 'popoverPosition'> {
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
            popoverPosition={position}
        >
            {children}
        </ReachMenuPopover>
    )
}) as ForwardReferenceComponent<'div', MenuListProps>

export interface PopoverProps extends PopoverContentProps {
    /**
     * Since ReachMenuPopover also has a prop that's named 'position' in order to
     * pass it prop properly to the as={Component} Component we have to
     * have unique prop to avoid prop name conflicts.
     */
    popoverPosition: Position
}

const Popover = React.forwardRef(({ popoverPosition, ...props }, reference) => (
    <PopoverContent
        {...props}
        as={MenuItems}
        ref={reference}
        position={popoverPosition}
        focusLocked={false}
        className={classNames('py-1', props.className)}
    />
)) as ForwardReferenceComponent<'div', PopoverProps>
