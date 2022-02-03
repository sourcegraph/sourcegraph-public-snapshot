import { Menu as ReachMenu, MenuProps as ReachMenuProps } from '@reach/menu-button'
import { isFunction, noop } from 'lodash'
import React, { ComponentType, forwardRef } from 'react'

import { ForwardReferenceComponent } from '../..'
import { Popover, PopoverProps } from '../Popover'

export type MenuProps = ReachMenuProps &
    PopoverProps & {
        as?: ComponentType
    }

/**
 * A Menu component.
 *
 * This component should be used to render an application menu that
 * presents a list of selectable items to the user.
 *
 * @see — Building accessible menus: https://www.w3.org/TR/wai-aria-practices/examples/menu-button/menu-button-links.html
 * @see — Docs https://reach.tech/menu-button#menu
 */
export const Menu = forwardRef((props, reference) => {
    const { children, isOpen, onOpenChange, as: Component, ...rest } = props
    const isControlled = isOpen !== undefined

    if (isControlled) {
        return (
            <ReachMenu as={Component} ref={reference} {...rest}>
                <Popover isOpen={isOpen} onOpenChange={onOpenChange ?? noop}>
                    {isFunction(children) ? children({ isOpen, isExpanded: isOpen }) : children}
                </Popover>
            </ReachMenu>
        )
    }

    return (
        <ReachMenu as={Component} ref={reference} {...rest}>
            {({ isExpanded }) => (
                <Popover isOpen={isExpanded} onOpenChange={noop}>
                    {isFunction(children) ? children({ isExpanded, isOpen: isExpanded }) : children}
                </Popover>
            )}
        </ReachMenu>
    )
}) as ForwardReferenceComponent<'div', MenuProps>
