import React, { useMemo } from 'react'

import { MenuButton as ReachMenuButton } from '@reach/menu-button'
import { uniqueId } from 'lodash'

import { ForwardReferenceComponent } from '../../types'
import { Button, ButtonProps } from '../Button'
import { PopoverTrigger } from '../Popover'

export type MenuButtonProps = Omit<ButtonProps, 'as'>

/**
 * Wraps a styled Wildcard `<Button />` component that can
 * toggle the opening and closing of a dropdown menu.
 *
 * @see — Docs https://reach.tech/menu-button#menubutton
 */
export const MenuButton = React.forwardRef(({ children, id, ...props }, reference) => {
    // To fix rule: "duplicate-id-active"
    // Document has active elements with the same id attribute: menu-button--menu
    const uniqueMenuId = useMemo(() => id ?? uniqueId('menu-button-'), [id])

    // aria-controls="" to fix "aria-valid-attr-value" since current version of
    // Popover removes contents from DOM, it causes `aria-controls="menu-*"` invalid
    // since `menu-*` doesn't exit on DOM

    return (
        <ReachMenuButton ref={reference} as={PopoverTriggerButton} id={uniqueMenuId} {...props} aria-controls="">
            {children}
        </ReachMenuButton>
    )
}) as ForwardReferenceComponent<'button', MenuButtonProps>

const PopoverTriggerButton = React.forwardRef((props, reference) => (
    <PopoverTrigger ref={reference} as={Button} {...props} />
)) as ForwardReferenceComponent<'button', ButtonProps>
