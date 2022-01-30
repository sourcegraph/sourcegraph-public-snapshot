import { Menu as ReachMenu, MenuProps as ReachMenuProps } from '@reach/menu-button'
import { isFunction, noop } from 'lodash'
import React from 'react'

import { Popover, PopoverProps } from '../Popover'

export type MenuProps = ReachMenuProps & PopoverProps

/**
 * A Menu component.
 *
 * This component should be used to render an application menu that
 * presents a list of selectable items to the user.
 *
 * @see — Building accessible menus: https://www.w3.org/TR/wai-aria-practices/examples/menu-button/menu-button-links.html
 * @see — Docs https://reach.tech/menu-button#menu
 */
export const Menu: React.FunctionComponent<MenuProps> = ({ children, isOpen, onOpenChange, ...props }) => {
    const isControlled = isOpen !== undefined

    if (isControlled) {
        return (
            <ReachMenu {...props}>
                <Popover isOpen={isOpen} onOpenChange={onOpenChange ?? noop}>
                    {isFunction(children) ? children({ isOpen, isExpanded: isOpen }) : children}
                </Popover>
            </ReachMenu>
        )
    }

    return (
        <ReachMenu {...props}>
            {({ isExpanded }) => (
                <Popover isOpen={isExpanded} onOpenChange={noop}>
                    {isFunction(children) ? children({ isExpanded, isOpen: isExpanded }) : children}
                </Popover>
            )}
        </ReachMenu>
    )
}
