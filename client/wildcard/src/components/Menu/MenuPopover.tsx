import { MenuPopover as ReachMenuPopover, MenuPopoverProps as ReachMenuPopoverProps } from '@reach/menu-button'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'

export type MenuPopoverProps = ReachMenuPopoverProps

/**
 * A popover component that will render `<MenuItems />` conditionally
 * depending on the `<MenuButton />` toggle.
 *
 * @see — Docs https://reach.tech/menu-button#menupopover
 */
export const MenuPopover = React.forwardRef(({ children, className, ...props }, reference) => (
    <ReachMenuPopover ref={reference} {...props} className={className}>
        {children}
    </ReachMenuPopover>
)) as ForwardReferenceComponent<'div', MenuPopoverProps>
