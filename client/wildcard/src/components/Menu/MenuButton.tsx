import { type ReactNode, useMemo, forwardRef } from 'react'

import { MenuButton as ReachMenuButton } from '@reach/menu-button'
import { uniqueId } from 'lodash'

import type { ForwardReferenceComponent } from '../../types'
import { Button, type ButtonProps } from '../Button'
import { PopoverTrigger, type PopoverTriggerProps } from '../Popover'

export interface MenuButtonProps extends Omit<ButtonProps, 'children'>, PopoverTriggerProps {}

/**
 * Wraps a styled Wildcard `<Button />` component that can
 * toggle the opening and closing of a dropdown menu.
 *
 * @see — Docs https://reach.tech/menu-button#menubutton
 */
export const MenuButton = forwardRef(({ id, children, ...props }, reference) => {
    // To fix rule: "duplicate-id-active"
    // Document has active elements with the same id attribute: menu-button--menu
    const uniqueMenuId = useMemo(() => id ?? uniqueId('menu-button-'), [id])

    // We unset aria-controls as it causes accessibility issues if the Menu is not yet rendered in the DOM.
    // aria-controls has low support across screen readers so this shouldn't be an issue: https://github.com/w3c/aria/issues/995
    return (
        <ReachMenuButton
            ref={reference}
            as={PopoverTriggerButton}
            id={uniqueMenuId}
            aria-controls=""
            // Pass empty string in order to suppress TS issue - ReachMenuButton always
            // should be call with children props
            // eslint-disable-next-line react/no-children-prop
            children=""
            // Pass real children prop with separate special prop to the PopoverTriggerButton because
            // reach-ui enforces prop-types check and in order to avoid this internal check we are
            // avoiding passing real children with standard children prop
            childrenContent={children}
            {...props}
        />
    )
}) as ForwardReferenceComponent<'button', MenuButtonProps>

interface PopoverTriggerButtonProps extends Omit<ButtonProps, 'children'>, PopoverTriggerProps {
    childrenContent?: ReactNode | ((isOpen: boolean) => ReactNode)
}

const PopoverTriggerButton = forwardRef(({ childrenContent, ...otherProps }, reference) => (
    <PopoverTrigger ref={reference} as={Button} {...otherProps}>
        {childrenContent}
    </PopoverTrigger>
)) as ForwardReferenceComponent<'button', PopoverTriggerButtonProps>
