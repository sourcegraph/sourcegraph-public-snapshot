import { type ComponentType, forwardRef, useMemo } from 'react'

import { Menu as ReachMenu, type MenuProps as ReachMenuProps } from '@reach/menu-button'
import { isFunction, noop, uniqueId } from 'lodash'

import type { ForwardReferenceComponent } from '../..'
import { Popover } from '../Popover'

export type MenuProps = ReachMenuProps & {
    as?: ComponentType<React.PropsWithChildren<unknown>>
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
    const { children, as: Component, id, ...rest } = props
    // To fix Rule: "aria-valid-attr-value"
    // Invalid ARIA attribute value: aria-controls="menu--1"
    const uniqueAriaControlId = useMemo(() => id ?? uniqueId('menu-'), [id])

    return (
        <ReachMenu as={Component} ref={reference} id={uniqueAriaControlId} {...rest}>
            {({ isExpanded }) => (
                <Popover isOpen={isExpanded} onOpenChange={noop}>
                    {isFunction(children) ? children({ isExpanded, isOpen: isExpanded }) : children}
                </Popover>
            )}
        </ReachMenu>
    )
}) as ForwardReferenceComponent<'div', MenuProps>
