import { Menu as ReachMenu, MenuProps as ReachMenuProps } from '@reach/menu-button'
import { isFunction } from 'lodash'
import React from 'react'

import { Popover } from '../Popover'

export type MenuProps = ReachMenuProps

/**
 * A Menu component.
 *
 * This component should be used to render an application menu that
 * presents a list of selectable items to the user.
 *
 * @see — Building accessible menus: https://www.w3.org/TR/wai-aria-practices/examples/menu-button/menu-button-links.html
 * @see — Docs https://reach.tech/menu-button#menu
 */
export const Menu: React.FunctionComponent<ReachMenuProps> = ({ children, ...props }) => (
    <ReachMenu {...props}>
        {({ isExpanded }) => (
            <Popover open={isExpanded}>
                {isFunction(children) ? children({ isExpanded, isOpen: isExpanded }) : children}
            </Popover>
        )}
    </ReachMenu>
)
